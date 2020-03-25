package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

//URLJoin joins urls with '/'
func URLJoin(urls ...string) string {
	u, err := url.Parse(urls[0])
	if err != nil || u.Host == "" {
		return strings.Join(urls, "/")
	}
	u.Path = path.Join(u.Path, path.Join(urls[1:]...))
	return u.String()
}

//GetURLFromConfig retrieves URL from config and checks it
func GetURLFromConfig(name string) (string, error) {
	return validateConfigURL(cmdapp.Config.GetString(name), name)
}

//GetURLFromConfig retrieves URL from config and checks it
func validateConfigURL(urlStr, settingName string) (string, error) {
	if urlStr == "" {
		return "", errors.New("No " + settingName + " setting provided")
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "Can't parse url "+urlStr)
	}
	return url.String(), nil
}

//ErrWrongHTTPCall indicates failure due wrong http call
var ErrWrongHTTPCall = errors.New("Wrong http call")

//ValidateResponse returns error if code is not in [200, 299]
func ValidateResponse(resp *http.Response) error {
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		trimS := ""
		if len(bodyBytes) > 100 {
			bodyBytes = bodyBytes[:100]
			trimS = "..."
		}
		msg := fmt.Sprintf("Wrong response code from server. Code: %d\n%s",
			resp.StatusCode, string(bodyBytes)+trimS)
		if resp.StatusCode == 400 {
			return errors.Wrapf(ErrWrongHTTPCall, msg)
		}
		return errors.New(msg)
	}
	return nil
}

//URLToLog removes pass from URL
func URLToLog(link string) string {
	u, err := url.Parse(link)
	if err == nil {
		if u.User != nil {
			u.User = url.UserPassword(u.User.Username(), "xxxx")
		}
		return u.String()
	}
	return link
}
