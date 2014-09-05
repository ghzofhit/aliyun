package aliyun

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestClient(t *testing.T) {

	c := New("ruCnke7MVynL9T11", "itWxg1RBZBiY2jVWo0wu9oOjjI7BVg")
	bucket := &Bucket{"8mbang-app-img", L_Beijing, c}
	content, err := os.Open("E:/Document/Desktop/News/005Hg45Vgw1ejxcj9hjr0j30c88351ky.jpg")
	err = bucket.PutObject("hello.jpg", "image/jpeg", content, map[string]string{"X-OSS-meta-test": "helloworld", "Cache-Control": "no-cache"})
	Convey("When have no entries.", t, func() {
		So(err, ShouldBeNil)
	})
	//t.Log(c.DeleteBucket("wnd", L_Beijing))
	// bucket, err := c.PutBucket("rsaa-cn", L_Beijing, ACL_Public_RDONLY)
	// t.Log(bucket, err)
	// content, err := os.Open("/One/Pictory/20111006687.jpg")
	// if err != nil {
	// 	t.Log(err)
	// 	return
	// }
	// err = bucket.PutObject("hello.jpg", "image/jpeg", content, map[string]string{"X-OSS-meta-test": "helloworld", "Cache-Control": "no-cache"})
	// t.Log(err)

	// err = bucket.DeleteObject("hello.jpg")
	// t.Log(err)
}
