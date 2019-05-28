package twitterdownloader

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Twitter struct {
	guest_token string
	playbackUrl string
	c           *Client
}

// Download jpeg from Twitter
func (t *Twitter) DownloadJPG(url string) (string, error) {
	if t.c.client == nil {
		// return errors.New("Client is nil, should init with proxy")
		t.c.client = &http.Client{}
	}
	return t.c.GetWithProxy(url)
}

// Download Video from Twitter by Guest
func (t *Twitter) DownloadVideo(url string) (string, error) {
	//Parse url to get tweet Id
	//https: //twitter.com/i/status/1035056498307522560
	uris := strings.Split(url, "/")
	var configJson string
	if len(uris) >= 5 {
		configJson = "https://api.twitter.com/1.1/videos/tweet/config/" + uris[5] + ".json"
	} else {
		return "", errors.New("URL provided shoud have form like (https: //twitter.com/i/status/1035056498307522560)")
	}

	err := t.activate(t.c)
	if err != nil {
		return "", err
	}
	err = t.activate2(t.c)
	if err != nil {
		return "", err
	}
	err = t.getVideoJson(t.c, configJson)
	if err != nil {
		return "", err
	}
	//jsonUrl := "https://api.twitter.com/1.1/videos/tweet/config/1035056498307522560.json"
	err = t.getVideoJson2(t.c, configJson)
	if err != nil {
		return "", err
	}
	//jsonUrl := "https://api.twitter.com/1.1/videos/tweet/config/1035056498307522560.json"
	name, err := t.getm3u8List(t.c)
	if err != nil {
		return "", err
	}
	return name, err
}

func (t *Twitter) SetupClient(c *Client) {
	t.c = c
}

func (t *Twitter) activate(c *Client) error {
	if c.client == nil {
		// return errors.New("Client is nil, should init with proxy")
		c.client = &http.Client{}
	}

	activateUrl := "https://api.twitter.com/1.1/guest/activate.json"
	req, err := http.NewRequest("OPTIONS", activateUrl, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")
	req.Header.Add("Access-Control-Request-Headers", "authorization,x-csrf-token")
	req.Header.Add("Access-Control-Request-Method", "POST")
	req.Header.Add("Origin", "https://twitter.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//fmt.Println("Activate Response: ", resp)
	return nil
}

func (t *Twitter) activate2(c *Client) error {
	if c.client == nil {
		c.client = &http.Client{}
	}
	activateUrl := "https://api.twitter.com/1.1/guest/activate.json"
	req, err := http.NewRequest("POST", activateUrl, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")
	req.Header.Add("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAIK1zgAAAAAA2tUWuhGZ2JceoId5GwYWU5GspY4%3DUq7gzFoCZs1QfwGoVdvSac3IniczZEYXIcDyumCauIXpcAPorE")
	req.Header.Add("cookie", `personalization_id="v1_pid1UUVchOmH31FJFT2ZLA=="; guest_id=v1%3A153569557385510737`)
	req.Header.Add("content-length", "0")
	req.Header.Add("Origin", "https://twitter.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	defer resp.Body.Close()

	var reader io.ReadCloser

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}
	body, err := ioutil.ReadAll(reader)

	var guest GuestToken
	err = json.Unmarshal(body, &guest)
	t.guest_token = guest.GuestToken
	return nil
}

func (t *Twitter) getVideoJson(c *Client, jsonUrl string) error {
	if c.client == nil {
		c.client = &http.Client{}
	}

	//jsonUrl := "https://api.twitter.com/1.1/videos/tweet/config/1035056498307522560.json"
	req, err := http.NewRequest("OPTIONS", jsonUrl, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")
	req.Header.Add("Access-Control-Request-Headers", "authorization,x-csrf-token")
	req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Origin", "https://twitter.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	defer resp.Body.Close()

	//fmt.Println("JsonConf Response: ", resp)
	return nil
}

func (t *Twitter) getVideoJson2(c *Client, jsonUrl string) error {
	if c.client == nil {
		c.client = &http.Client{}
	}
	//jsonUrl := "https://api.twitter.com/1.1/videos/tweet/config/1035056498307522560.json"
	req, err := http.NewRequest("GET", jsonUrl, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")
	req.Header.Add("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAIK1zgAAAAAA2tUWuhGZ2JceoId5GwYWU5GspY4%3DUq7gzFoCZs1QfwGoVdvSac3IniczZEYXIcDyumCauIXpcAPorE")
	req.Header.Add("cookie", `personalization_id="v1_pid1UUVchOmH31FJFT2ZLA=="; guest_id=v1%3A153569557385510737`)
	req.Header.Add("content-length", "0")
	req.Header.Add("Origin", "https://twitter.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	req.Header.Add("x-guest-token", t.guest_token)

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	defer resp.Body.Close()

	var reader io.ReadCloser

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}
	body, err := ioutil.ReadAll(reader)

	var vc VideoConfig
	err = json.Unmarshal(body, &vc)
	//fmt.Println("respJons:", vc, err)
	t.playbackUrl = vc.Track.PlaybackUrl
	return nil
}

func (t *Twitter) getm3u8List(c *Client) (string, error) {
	if c.client == nil {
		c.client = &http.Client{}
	}

	req, err := http.NewRequest("GET", t.playbackUrl, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")
	// req.Header.Add("Access-Control-Request-Headers", "authorization,x-csrf-token")
	// req.Header.Add("Access-Control-Request-Method", "GET")
	req.Header.Add("Origin", "https://twitter.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// fmt.Println("Getm3u8List:", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	var reader io.ReadCloser

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	//Trim ?tag=5
	videoUrl := strings.Split(t.playbackUrl, "?")[0]
	if strings.HasSuffix(videoUrl, ".mp4") {
		filename := extractFilename(videoUrl)
		saveFile(filename, reader)
		return "", nil
	}
	uri, err := playList(reader)
	if err != nil {
		return "", err
	}

	fileName := extractFilename(t.playbackUrl)
	fileName = strings.Split(fileName, ".")[0] + ".mp4"
	m3u8Url := "https://video.twimg.com" + uri

	m3FileName, err := saveM3U8(m3u8Url, c)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("ffmpeg", `-i`, m3FileName, "-codec", "copy", "-bsf:a", "aac_adtstoasc", fileName)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		os.Remove(m3FileName)
		return "", errors.New(fmt.Sprint(err) + ": " + stderr.String())
	}
	os.Remove(m3FileName)
	return fileName, err
}

func saveM3U8(url string, c *Client) (fileName string, err error) {
	if c.client == nil {
		c.client = &http.Client{}
	}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")

	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var reader io.ReadCloser

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}
	b, _ := ioutil.ReadAll(reader)
	str := string(b)
	r := regexp.MustCompile(`\/ext_tw_video\/\d+\/.+.ts`)
	urls := r.FindAllString(str, -1)
	for _, url := range urls {
		newUrl := strings.Replace(url, url, "https://video.twimg.com"+url, -1)
		str = strings.Replace(str, url, newUrl, -1)
	}
	b = []byte(str)
	fileName = bson.NewObjectId().Hex() + ".m3u8"
	err = ioutil.WriteFile(fileName, b, os.ModePerm)
	return
}
