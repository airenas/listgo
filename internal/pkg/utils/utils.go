package utils;

import (
	"github.com/pkg/errors"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"strings"
	"path"
	"net/url"
)

//URLJoin joins urls with '/'
func URLJoin(urls ...string) string {
	u, err := url.Parse(urls[0])
	if (err != nil || u.Host == ""){
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