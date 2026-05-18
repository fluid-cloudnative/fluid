package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type listBucketResult struct {
	XMLName   xml.Name      `xml:"ListBucketResult"`
	Name      string        `xml:"Name"`
	Prefix    string        `xml:"Prefix,omitempty"`
	Delimiter string        `xml:"Delimiter,omitempty"`
	Marker    string        `xml:"Marker,omitempty"`
	MaxKeys   int           `xml:"MaxKeys"`
	KeyCount  int           `xml:"KeyCount,omitempty"`
	IsTrunc   bool          `xml:"IsTruncated"`
	Contents  []objectEntry `xml:"Contents"`
}

type objectEntry struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Type         string `xml:"Type"`
	Size         int    `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

type errorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId,omitempty"`
	HostID    string   `xml:"HostId,omitempty"`
}

const emulatorLastModified = "Tue, 20 Apr 2026 00:00:00 GMT"

func main() {
	bucketName := getenv("BUCKET_NAME", "bucket-a")
	objectKey := getenv("OBJECT_KEY", "testfile")
	objectAlias := getenv("OBJECT_ALIAS", "")
	objectValue := getenv("OBJECT_VALUE", "bucket-a-data")
	accessKeyID := getenv("ACCESS_KEY_ID", "")
	addr := getenv("LISTEN_ADDR", ":9000")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("request method=%s host=%s path=%s rawQuery=%s", r.Method, r.Host, r.URL.Path, r.URL.RawQuery)

		if !authorizeRequest(w, r, accessKeyID) {
			return
		}

		if r.URL.Query().Get("location") != "" {
			w.Header().Set("Content-Type", "application/xml")
			fmt.Fprintf(w, "<LocationConstraint>oss-cn-hangzhou</LocationConstraint>")
			return
		}

		if r.Method == http.MethodHead || r.Method == http.MethodGet {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusOK)
					return
				}

				prefix := strings.TrimPrefix(r.URL.Query().Get("prefix"), "/")
				if prefix != "" && !strings.HasSuffix(prefix, "/") {
					prefix += "/"
				}

				var contents []objectEntry
				if strings.HasPrefix(objectKey, prefix) || prefix == "" {
					contents = append(contents, objectEntry{
						Key:          objectKey,
						LastModified: "2026-04-20T00:00:00.000Z",
						ETag:         "\"dummy-etag\"",
						Type:         "Normal",
						Size:         len(objectValue),
						StorageClass: "Standard",
					})
				}
				if objectAlias != "" && objectAlias != objectKey && (strings.HasPrefix(objectAlias, prefix) || prefix == "") {
					contents = append(contents, objectEntry{
						Key:          objectAlias,
						LastModified: "2026-04-20T00:00:00.000Z",
						ETag:         "\"dummy-etag\"",
						Type:         "Normal",
						Size:         len(objectValue),
						StorageClass: "Standard",
					})
				}

				w.Header().Set("Content-Type", "application/xml")
				if err := xml.NewEncoder(w).Encode(listBucketResult{
					Name:      bucketName,
					Prefix:    r.URL.Query().Get("prefix"),
					Delimiter: r.URL.Query().Get("delimiter"),
					MaxKeys:   1000,
					KeyCount:  len(contents),
					IsTrunc:   false,
					Contents:  contents,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			if path == objectKey || path == objectAlias {
				writeObjectHeaders(w.Header(), len(objectValue), "Normal")
				if r.Method == http.MethodGet {
					_, _ = w.Write([]byte(objectValue))
				}
				return
			}

			trimmedObjectDir := strings.TrimSuffix(objectKey, "/")
			objectDir := trimmedObjectDir
			if idx := strings.LastIndex(trimmedObjectDir, "/"); idx >= 0 {
				objectDir = trimmedObjectDir[:idx+1]
			}
			if path == objectDir || path == strings.TrimSuffix(objectDir, "/") {
				writeObjectHeaders(w.Header(), 0, "Directory")
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.Header().Set("Content-Type", "application/xml")
				if err := xml.NewEncoder(w).Encode(listBucketResult{
					Name:      bucketName,
					Prefix:    objectDir,
					Delimiter: r.URL.Query().Get("delimiter"),
					MaxKeys:   1000,
					KeyCount:  1,
					IsTrunc:   false,
					Contents: []objectEntry{{
						Key:          objectKey,
						LastModified: "2026-04-20T00:00:00.000Z",
						ETag:         "\"dummy-etag\"",
						Type:         "Normal",
						Size:         len(objectValue),
						StorageClass: "Standard",
					}},
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			http.NotFound(w, r)
			return
		}

		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	})

	log.Printf("starting local oss emulator for bucket=%s key=%s on %s", bucketName, objectKey, addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func writeObjectHeaders(header http.Header, contentLength int, objectType string) {
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Length", fmt.Sprintf("%d", contentLength))
	header.Set("Content-MD5", "1B2M2Y8AsgTpgAmY7PhCfg==")
	header.Set("Last-Modified", emulatorLastModified)
	header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	header.Set("Server", "AliyunOSS")
	header.Set("Accept-Ranges", "bytes")
	header.Set("ETag", "\"dummy-etag\"")
	header.Set("X-Oss-Request-Id", "oss-emulator")
	header.Set("X-Oss-Version-Id", "null")
	header.Set("X-Oss-Last-Access-Time", emulatorLastModified)
	header.Set("X-Oss-Object-Type", objectType)
	header.Set("X-Oss-Storage-Class", "Standard")
	header.Set("X-Oss-Hash-Crc64ecma", "123456789")
}

func authorizeRequest(w http.ResponseWriter, r *http.Request, expectedAccessKeyID string) bool {
	if expectedAccessKeyID == "" {
		return true
	}

	accessKeyID := extractAccessKeyID(r)
	if accessKeyID == expectedAccessKeyID {
		return true
	}

	log.Printf(
		"reject unauthorized request: invalid access key id method=%s host=%s path=%s key_present=%t",
		r.Method,
		r.Host,
		r.URL.Path,
		accessKeyID != "",
	)
	writeAuthError(w, "InvalidAccessKeyId", "The OSS emulator rejected the provided AccessKeyId.")
	return false
}

func extractAccessKeyID(r *http.Request) string {
	if accessKeyID := r.URL.Query().Get("OSSAccessKeyId"); accessKeyID != "" {
		return accessKeyID
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(authHeader, "OSS ") {
		credential := strings.TrimPrefix(authHeader, "OSS ")
		accessKeyID, _, _ := strings.Cut(credential, ":")
		return strings.TrimSpace(accessKeyID)
	}

	if strings.HasPrefix(authHeader, "OSS4-HMAC-SHA256 ") {
		for _, part := range strings.Split(strings.TrimPrefix(authHeader, "OSS4-HMAC-SHA256 "), ",") {
			part = strings.TrimSpace(part)
			if !strings.HasPrefix(part, "Credential=") {
				continue
			}
			credential := strings.TrimPrefix(part, "Credential=")
			accessKeyID, _, _ := strings.Cut(credential, "/")
			return strings.TrimSpace(accessKeyID)
		}
	}

	return ""
}

func writeAuthError(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusForbidden)
	if err := xml.NewEncoder(w).Encode(errorResponse{
		Code:      code,
		Message:   message,
		RequestID: "oss-emulator",
		HostID:    "oss-emulator",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
