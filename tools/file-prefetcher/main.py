import os
import glob
import time

buffer_size_in_bytes = int(os.getenv("BUFFER_SIZE_IN_BYTES", "16777216"))# 16MiB

def file_read(file):
    print("")
    buffer_size = buffer_size_in_bytes
    with open(file, 'rb') as f:
        while True:
            buffer = f.read(buffer_size)
            if not buffer:
                break

def main():
    glob_patterns = os.getenv("FILE_PREFETCHER_FILE_LIST", None)
    assert glob_patterns is not None, "env variable FILE_PREFETCHER_FILE_LIST is not set"

     # Get all files to prefetch
    files_to_prefetch = []
    glob_patterns = glob_patterns.split(";")
    for glob_pattern in glob_patterns:
        if os.path.isdir(glob_pattern):
            glob_pattern = os.path.join(glob_pattern, "**")
        files = glob.glob(glob_pattern, recursive=True)
        files_to_prefetch.extend([file for file in files if os.path.isfile(file)])

    print(f"Found {len(files_to_prefetch)} files to prefetch")
    start_time = time.time()
    for file in files_to_prefetch:
        file_start_time = time.time()
        file_read(file)
        print(f"Prefetching file {file} end in {time.time() - file_start_time:.2f} seconds")
    print(f"Total time: {time.time() - start_time:.2f} seconds")

if __name__ == '__main__':
    prefetch_result = ""
    try:
        main()
        prefetch_result = "success"
    except Exception as e:
        print(e)
        prefetch_result = "fail"
    finally:
        os.makedirs("/tmp/fluid-file-prefetcher/status/", exist_ok=True)
        with open("/tmp/fluid-file-prefetcher/status/prefetcher.status", "w") as f:
            f.write(f"prefetch_result={prefetch_result}\n")