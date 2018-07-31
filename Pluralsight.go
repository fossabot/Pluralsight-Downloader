package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type authStruct struct {
	DeviceID     string `json:"deviceId"`
	RefreshToken string `json:"refreshToken"`
}

type tokenStruct struct {
	Token         string    `json:"token"`
	Jwt           string    `json:"jwt"`
	UserHandle    string    `json:"userHandle"`
	Expiration    time.Time `json:"expiration"`
	Authenticated bool      `json:"authenticated"`
	OfflineAccess struct {
		ViewOffline bool `json:"viewOffline"`
		CacheDays   int  `json:"cacheDays"`
		CourseLimit int  `json:"courseLimit"`
	} `json:"offlineAccess"`
}

type videoMetaDataStruct struct {
	Data struct {
		ViewClip struct {
			Urls []struct {
				URL    string      `json:"url"`
				Cdn    string      `json:"cdn"`
				Rank   int         `json:"rank"`
				Source interface{} `json:"source"`
			} `json:"urls"`
			Status int `json:"status"`
		} `json:"viewClip"`
	} `json:"data"`
}

type playListsStruct struct {
	Data struct {
		RPC struct {
			BootstrapPlayer struct {
				Course struct {
					Name                           string `json:"name"`
					Title                          string `json:"title"`
					CourseHasCaptions              bool   `json:"courseHasCaptions"`
					SupportsWideScreenVideoFormats bool   `json:"supportsWideScreenVideoFormats"`
					Timestamp                      string `json:"timestamp"`
					Modules                        []struct {
						Name              string `json:"name"`
						Title             string `json:"title"`
						Duration          int    `json:"duration"`
						FormattedDuration string `json:"formattedDuration"`
						Author            string `json:"author"`
						Authorized        bool   `json:"authorized"`
						Clips             []struct {
							Authorized        bool   `json:"authorized"`
							ClipID            string `json:"clipId"`
							Duration          int    `json:"duration"`
							FormattedDuration string `json:"formattedDuration"`
							ID                string `json:"id"`
							Index             int    `json:"index"`
							ModuleIndex       int    `json:"moduleIndex"`
							ModuleTitle       string `json:"moduleTitle"`
							Name              string `json:"name"`
							Title             string `json:"title"`
							Watched           bool   `json:"watched"`
						} `json:"clips"`
					} `json:"modules"`
				} `json:"course"`
			} `json:"bootstrapPlayer"`
		} `json:"rpc"`
	} `json:"data"`
}

func OneGetAuth(email string, password string) *authStruct {
	var cdnurl = "https://app.pluralsight.com/mobile-api/v2/user/device/authenticated"
	client := &http.Client{}
	query := "{\"DeviceModel\":\"Windows Desktop\",\"DeviceName\":null,\"Username\":\"" + email + "\",\"Password\":\"" + password + "\"}"
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", cdnurl, bytes.NewBuffer(jsonStr))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	resp, _ := client.Do(req)
	if resp.StatusCode != 200 {
		fmt.Println("Authorization Failed")
		os.Exit(1)
	}
	_authStruct := new(authStruct)
	json.NewDecoder(resp.Body).Decode(_authStruct)
	// fmt.Println(_authStruct.DeviceID)
	defer resp.Body.Close()
	return _authStruct
}

func TwoGetToken(_authStruct *authStruct) *tokenStruct {
	var cdnurl = "https://app.pluralsight.com/mobile-api/v2/user/authorization/" + _authStruct.DeviceID
	client := &http.Client{}
	_authStructJson, _ := json.Marshal(_authStruct)
	var jsonStr = []byte(_authStructJson)
	req, _ := http.NewRequest("POST", cdnurl, bytes.NewBuffer(jsonStr))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	resp, _ := client.Do(req)
	_tokenStruct := new(tokenStruct)
	json.NewDecoder(resp.Body).Decode(_tokenStruct)
	// fmt.Println(_tokenStruct.Jwt)
	defer resp.Body.Close()
	return _tokenStruct
}

func ThreeGetPlayLists(_tokenStruct *tokenStruct, course_name string) *playListsStruct {
	var cdnurl = "https://app.pluralsight.com/player/api/graphql"
	var _query = "{\"query\":\"query BootstrapPlayer {rpc{bootstrapPlayer{course(courseId: \\\"" + course_name + "\\\") {        name title courseHasCaptions   supportsWideScreenVideoFormats timestamp modules {   name   title   duration   formattedDuration   author   authorized   clips {     authorized     clipId     duration     formattedDuration     id     index     moduleIndex     moduleTitle     name     title      watched }}}}}}\"}"
	// fmt.Println(_query)
	client := &http.Client{}
	var jsonStr = []byte(_query)
	req, _ := http.NewRequest("POST", cdnurl, bytes.NewBuffer(jsonStr))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", "PsJwt-production="+_tokenStruct.Jwt+";")
	resp, _ := client.Do(req)
	_playListsStruct := new(playListsStruct)
	json.NewDecoder(resp.Body).Decode(_playListsStruct)
	defer resp.Body.Close()
	return _playListsStruct
}

func FourGetVideoMetadata(_tokenStruct *tokenStruct, id string) string {
	time.Sleep(1 * time.Second)
	var cdnurl = "https://app.pluralsight.com/player/api/graphql"
	client := &http.Client{}
	//"go-cli-playbook:45c18713-161d-4455-a309-56507c151581:3:mike-vansickle"
	course_name := strings.Split(id, ":")[0]
	clip_id := strings.Split(id, ":")[1]
	clip_index := strings.Split(id, ":")[2]
	author := strings.Split(id, ":")[3]
	query := "{\"query\":\" query viewClip {  viewClip(input: {    author: \\\"" + author + "\\\",     clipIndex: " + clip_index + ",     courseName: \\\"" + course_name + "\\\",     includeCaptions: false,     locale: \\\"en\\\",     mediaType: \\\"mp4\\\",     moduleName: \\\"" + clip_id + "\\\",     quality: \\\"1280x720\\\"  }) {    urls {      url      cdn      rank      source    },    status  }        }      \"}"
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", cdnurl, bytes.NewBuffer(jsonStr))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	req.Header.Set("cookie", "PsJwt-production="+_tokenStruct.Jwt+";")
	resp, _ := client.Do(req)
	_videoMetaData := new(videoMetaDataStruct)
	json.NewDecoder(resp.Body).Decode(_videoMetaData)
	//fmt.Println("Status Code: ", resp.StatusCode)
	//fmt.Println(_videoMetaData.Data.ViewClip.Urls)
	if _videoMetaData.Data.ViewClip.Status == 429 {
		fmt.Println("Status Code: ", _videoMetaData.Data.ViewClip)
		htmlData, err := ioutil.ReadAll(resp.Body) //<--- here!
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(htmlData)) //<-- here !
		return strconv.Itoa(_videoMetaData.Data.ViewClip.Status) + " status"
	}
	defer resp.Body.Close()
	return _videoMetaData.Data.ViewClip.Urls[0].URL
}

func worker(i int, jobs <-chan list, results chan<- string) {
	for j := range jobs {
		// fmt.Println("worker", i, "started  job", j.id)
		client := &http.Client{}
		req, _ := http.NewRequest("GET", j.url, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
		resp, _ := client.Do(req)
		// fmt.Println(resp.StatusCode)
		defer resp.Body.Close()
		reader, _ := ioutil.ReadAll(resp.Body)
		ioutil.WriteFile(j.fileName, []byte(string(reader)), 0x777) // Write to the file i as a byte array
		resp.Body.Close()
		//fmt.Println("worker", i, "finished job", j.id)
		results <- j.fileName
	}
}

type list struct {
	fileName string `json:"fileName"`
	url      string `json:"url"`
}

var name = flag.String("name", "", "Enter the username")
var password = flag.String("password", "", "Enter the password")
var playlisturl = flag.String("playlisturl", "", "Enter the playlist url")

func main() {
	flag.Parse()
	_authStruct := OneGetAuth(*name, *password)
	u, _ := url.Parse(*playlisturl)
	fragments, _ := url.ParseQuery(u.RawQuery)
	_tokenStruct := TwoGetToken(_authStruct)
	_playListsStruct := ThreeGetPlayLists(_tokenStruct, fragments["course"][0])
	modules := _playListsStruct.Data.RPC.BootstrapPlayer.Course.Modules
	_list := make([]list, 0)
	for i := 0; i < len(modules); i++ {
		module := modules[i]
		for j := 0; j < len(module.Clips); j++ {
			//fmt.Println(module.Clips[j])
			name := strconv.Itoa(i) + "_" + strings.Replace(module.Title, " ", "_", -1) + "_" + strconv.Itoa(j) + "_" + strings.Replace(module.Clips[j].Title, " ", "_", -1) + ".mp4"
			name = strings.Replace(name, ":", "", -1)
			name = strings.Replace(name, "'", "", -1)
			url := FourGetVideoMetadata(_tokenStruct, module.Clips[j].ID)
			fmt.Println("Processing MetaData for:", name)
			_list = append(_list, list{fileName: name, url: url})
			// break
		}
		// break
	}
	workers := 2
	fmt.Println("Workers: ", workers)
	length := len(_list)
	jobs := make(chan list)
	results := make(chan string, length+1)
	for i := 0; i < workers; i++ {
		go worker(i, jobs, results)
	}
	start := time.Now()
	for i := 0; i < length; i++ {
		jobs <- _list[i]
	}
	close(jobs)
	// Finally we collect all the results of the work.
	for a := 0; a < length; a++ {
		fmt.Println("Download : ", <-results)
	}
	fmt.Println("Time Taken: ", time.Since(start).String())
}

// fmt.Println("Status Code: ", resp.StatusCode)
// htmlData, err := ioutil.ReadAll(resp.Body) //<--- here!
// if err != nil {
// 	fmt.Println(err)
// 	os.Exit(1)
// }
// fmt.Println(string(htmlData)) //<-- here !

// listJson, _ := json.Marshal(_list)
// ioutil.WriteFile(_playListsStruct.Data.RPC.BootstrapPlayer.Course.Name, listJson, 0644)

// values := map[string]string{"username": username, "password": password}
// jsonValue, _ := json.Marshal(values)
// resp, err := http.Post(authAuthenticatorUrl, "application/json", bytes.NewBuffer(jsonValue))
