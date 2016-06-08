package nerve

//import (
//	"net/http"
//	"github.com/n0rad/go-erlog/errs"
//	"net/url"
//)
//
////type Policy bool
////const PolicyAll Policy = 0
////const PolicyAny Policy = 1
//
//type CheckHttpProxy struct {
//	CheckCommon
//	ProxyHost     string
//	ProxyPort     int
//	ProxyUsername string
//	ProxyPassword string
//	Urls          []string
//	//Policy   Policy
//
//	client        http.Client
//}
//
//func NewCheckHttpProxy() *CheckHttpProxy {
//	return &CheckHttpProxy{}
//}
//
//func (x *CheckHttpProxy) Init(conf *Service) error {
//	proxyUrl := "http://"
//	user := url.User(x.ProxyUsername)
//	user := url.UserPassword(x.ProxyUsername, x.ProxyPassword)
//
//	if
//	url.Parse("http://" + x.ProxyHost + ":" + x.ProxyPort)
//
//	proxy := url.URL
//	x.client = http.Client{
//		Transport: &http.Transport{
//			Proxy: http.ProxyURL(x.ProxyUrl),
//		},
//		Timeout: x.TimeoutInMilli,
//	}
//
//	return nil
//}
//
//func (x *CheckHttpProxy) Check() (CheckStatus, error) {
//	for _, url := range x.Urls {
//		resp, err := x.client.Get(url)
//		if err != nil {
//			return KO, errs.WithEF(err, x.fields.WithField("url", url), "Url check failed")
//		}
//		resp.Body.Close()
//	}
//
//	return OK, nil
//}
