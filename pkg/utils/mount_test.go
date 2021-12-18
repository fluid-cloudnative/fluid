package utils

import (
	"errors"
	. "github.com/agiledragon/gomonkey"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"k8s.io/utils/mount"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestMountRootWithEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/var/lib/mymount"},
	}
	for _, tc := range testCases {
		os.Setenv(MountRoot, tc.input)
		mountRoot, err := GetMountRoot()
		if err != nil {
			t.Errorf("Get error %v", err)
		}
		if tc.expected != mountRoot {
			t.Errorf("expected %#v, got %#v",
				tc.expected, mountRoot)
		}
	}
}

func TestMountRootWithoutEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", ""},
	}
	for _, tc := range testCases {
		os.Unsetenv(MountRoot)
		mountRoot, err := GetMountRoot()
		if err == nil {
			t.Errorf("Expected error happened, but no error")
		}

		if err.Error() != "the the value of the env variable named MOUNT_ROOT is illegal" {
			t.Errorf("Get unexpected error %v", err)
		}

		if tc.expected != mountRoot {
			t.Errorf("Unexpected result %s", tc.expected)
		}

	}
}

func TestCheckMountReady(t *testing.T) {
	Convey("TestCheckMountReady", t, func() {
		Convey("CheckMountReady success", func() {
			cmd := &exec.Cmd{}
			patch1 := ApplyMethod(reflect.TypeOf(cmd), "CombinedOutput", func(_ *exec.Cmd) ([]byte, error) {
				return nil, nil
			})
			defer patch1.Reset()

			err := CheckMountReady("/test", "test")
			So(err, ShouldBeNil)
		})
		Convey("CheckMountReady false", func() {
			cmd := &exec.Cmd{}
			patch1 := ApplyMethod(reflect.TypeOf(cmd), "CombinedOutput", func(_ *exec.Cmd) ([]byte, error) {
				return nil, errors.New("test")
			})
			defer patch1.Reset()

			err := CheckMountReady("/test", "test")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestIsMounted(t *testing.T) {
	Convey("TestIsMounted", t, func() {
		Convey("IsMounted success", func() {
			patch2 := ApplyFunc(os.Stat, func(filename string) (os.FileInfo, error) {
				return nil, nil
			})
			defer patch2.Reset()
			patch1 := ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio /var/lib/kubelet/pods/4781fc5b-72f9-4175-9321-2e1f169880ce/volumes/kubernetes.io~csi/default-jfsdemo/mount fuse.juicefs rw,relatime,user_id=0,group_id=0,default_permissions,allow_other 0 0"), nil
			})
			defer patch1.Reset()
			absPath := "/var/lib/kubelet/pods/4781fc5b-72f9-4175-9321-2e1f169880ce/volumes/kubernetes.io~csi/default-jfsdemo/mount"

			mounted, err := IsMounted(absPath)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeTrue)
		})
		Convey("IsMounted false", func() {
			patch1 := ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio /var/lib/kubelet/pods/4781fc5b-72f9-4175-9321-2e1f169880ce/volumes/kubernetes.io~csi/default-jfsdemo/mount fuse.juicefs rw,relatime,user_id=0,group_id=0,default_permissions,allow_other 0 0"), nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(os.Stat, func(filename string) (os.FileInfo, error) {
				return nil, nil
			})
			defer patch2.Reset()
			absPath := "/test"

			mounted, err := IsMounted(absPath)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeFalse)
		})
		Convey("IsMounted error", func() {
			patch1 := ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio"), errors.New("test")
			})
			defer patch1.Reset()
			absPath := "/test"

			mounted, err := IsMounted(absPath)
			So(err, ShouldNotBeNil)
			So(mounted, ShouldBeFalse)
		})
	})
}

func TestCheckMountPointBroken(t *testing.T) {
	Convey("TestCheckMountPointBroken", t, func() {
		Convey("CheckMountPointBroken success", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return true, errors.New("test")
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mount.IsCorruptedMnt, func(err error) bool {
				return true
			})
			defer patch2.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldBeNil)
			So(broken, ShouldBeTrue)
		})
		Convey("CheckMountPointBroken not broken", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return true, nil
			})
			defer patch1.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldBeNil)
			So(broken, ShouldBeFalse)
		})
		Convey("CheckMountPointBroken not exist", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return false, nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mount.IsCorruptedMnt, func(err error) bool {
				return false
			})
			defer patch2.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldNotBeNil)
			So(broken, ShouldBeFalse)
		})
		Convey("CheckMountPointBroken error", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return false, errors.New("test")
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mount.IsCorruptedMnt, func(err error) bool {
				return false
			})
			defer patch2.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldNotBeNil)
			So(broken, ShouldBeFalse)
		})
		Convey("CheckMountPointBroken nil", func() {
			broken, err := CheckMountPointBroken("")
			So(err, ShouldNotBeNil)
			So(broken, ShouldBeFalse)
		})
	})
}
