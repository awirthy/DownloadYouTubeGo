package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type settings struct {
	Email             string
	MediaFolder       string
	MediaFolderNotify string
	RSSFolder         string
	RSSTemplate       string
	HTTPHost          string
	Config            string
	PlaylistItems     string
	PushoverUserToken string
	PodcastDownload   []YouTubeDownload `xml:"PodcastDownload"`
	PodcastsNotifty   []PodcastsNotifty `xml:"PodcastsNotifty"`
	RSSDownload       []RSSDownload     `xml:"RSSDownload"`
}

// type Tiktokfeed struct {
// 	TikTok TikTok `xml:"feed"`
// }

type TikTok struct {
	Title   string
	ID      string
	Icon    string
	Logo    string
	Updated string
	Author  string
	Link    string
}

// type JsonItemsMap struct {
// 	Items map[string]map[string]interface{} `json:"items"`
// }

type Entry struct {
	Title     string
	Published string
	Updated   string
	ID        string
	Link      string
	Content   string
}

type Validate struct {
	MediaFolder                     bool
	MediaFolderNotify               bool
	RSSFolder                       bool
	RSSTemplate                     bool
	Config                          bool
	PodcastDownload_Name            bool
	PodcastDownload_ChannelID       bool
	PodcastDownload_DownloadArchive bool
	PodcastDownload_FileFormat      bool
	PodcastDownload_FileQuality     bool
	PlaylistItems                   bool
	PushoverUserToken               bool
	PodcastDownload_YouTubeURL      bool
	PodcastsNotifty_Name            bool
	PodcastsNotifty_YouTubeURL      bool
	RSSDownload_Name                bool
	RSSDownload_ChannelID           bool
	RSSDownload_DownloadArchive     bool
	RSSDownload_FileFormat          bool
	RSSDownload_FileQuality         bool
	RSSDownload_PushoverAppToken    bool
	HTTPHost                        bool
	TikTokUsername                  bool
	TikTokFeed                      bool
	TikTok_PushoverAppToken         bool
}

type YouTubeDownload struct {
	Name             string `xml:"Name"`
	ChannelID        string `xml:"ChannelID"`
	FileFormat       string `xml:"FileFormat"`
	DownloadArchive  string `xml:"DownloadArchive"`
	FileQuality      string `xml:"FileQuality"`
	ChannelThumbnail string `xml:"ChannelThumbnail"`
	YouTubeURL       string `xml:"YouTubeURL"`
	PushoverAppToken string `xml:"PushoverAppToken"`
	// PushoverAppToken
}

// RSSDownload Name="jimmyrees (TikTok)" ChannelID="TikTok" TikTokUsername="jimmyrees" FileFormat="mp4" DownloadArchive="/config/youtube-dl-archive-TikTok-ALL.txt" FileQuality="best" ChannelThumbnail="https://www.tiktok.com/favicon.ico" TikTokFeed="http://10.0.0.186:3008/?action=display&amp;bridge=TikTokBridge&amp;format=Atom&amp;context=By+user&amp;username=%40" />
type RSSDownload struct {
	Name             string `xml:"Name"`
	ChannelID        string `xml:"ChannelID"`
	TikTokUsername   string `xml:"TikTokUsername"`
	FileFormat       string `xml:"FileFormat"`
	DownloadArchive  string `xml:"DownloadArchive"`
	FileQuality      string `xml:"FileQuality"`
	ChannelThumbnail string `xml:"ChannelThumbnail"`
	YouTubeURL       string `xml:"YouTubeURL"`
	TikTokFeed       string `xml:"TikTokFeed"`
	PushoverAppToken string `xml:"PushoverAppToken"`
}

type PodcastsNotifty struct {
	Name             string `xml:"Name"`
	YouTubeURL       string `xml:"YouTubeURL"`
	PushoverAppToken string `xml:"PushoverAppToken"`
}

type JsonData struct {
	id              string
	title           string
	webpage_url     string
	thumbnail       string
	description     string
	uploader_url    string
	channel_url     string
	duration_string string
	filesize_approx float64
}

type JsonChannelData struct {
	thumbnail   string
	description string
}

func isOlderThan(t time.Time) bool {
	return time.Now().Sub(t) > 168*time.Hour
}

func DeleteOldFiles(dir string) {
	descfiles, descerr := WalkMatch(dir, "*.description")

	if descerr != nil {
		log.Printf("------------------      START List description Files ERROR")
		log.Fatal(descerr)
		log.Printf("------------------      END List description Files ERROR")
	}

	for _, fname := range descfiles {
		arrfname_noext := strings.Split(fname, ".")
		fname_noext := arrfname_noext[0]

		log.Printf("fname:" + fname)
		log.Printf("fname_noext:" + fname_noext)

		fname_file, fname_fileerr := os.Stat(fname)

		if fname_fileerr != nil {
			log.Printf("------------------      START List fname_fileerr ERROR")
			log.Fatal(fname_fileerr)
			log.Printf("------------------      END List fname_fileerr ERROR")
		}

		if isOlderThan(fname_file.ModTime()) {
			log.Printf("DELETE FILE: " + fname_noext + ".description")
			os.Remove(fname_noext + ".description")
			log.Printf("DELETE FILE: " + fname_noext + ".mp4")
			os.Remove(fname_noext + ".mp4")
			log.Printf("DELETE FILE: " + fname_noext + ".info.json")
			os.Remove(fname_noext + ".info.json")
		}
	}
}

func IsValid(fp string) bool {
	// Check if file already exists
	if _, err := os.Stat(fp); err == nil {
		return true
	}
	return false
}

func handleJSONObject(object interface{}, key, indentation string) {
	switch t := object.(type) {
	case string:
		fmt.Println(indentation+key+": ", t) // t has type string
	case bool:
		fmt.Println(indentation+key+": ", t) // t has type bool
	case float64:
		fmt.Println(indentation+key+": ", t) // t has type float64 (which is the type used for all numeric types)
	case map[string]interface{}:
		fmt.Println(indentation + key + ":")
		for k, v := range t {
			handleJSONObject(v, k, indentation+"\t")
		}
	case []interface{}:
		fmt.Println(indentation + key + ":")
		for index, v := range t {
			handleJSONObject(v, "["+strconv.Itoa(index)+"]", indentation+"\t")
		}
	}
}

func IsValidURL(fp string) bool {
	log.Printf("URL Check: " + fp)

	resp, err := http.Get(fp)
	if err != nil {
		// print(err.Error())
		log.Printf("IsValidURL Status: " + resp.Status)
		// log.Printf("IsValidURL Error: " + err.Error())
		return false
	} else {
		if strings.Contains(resp.Status, "200 OK") {
			// print(string(resp.StatusCode) + resp.Status)
			log.Printf("URL Status: " + resp.Status)
			return true
		} else {
			// print(err.Error())
			log.Printf("IsValidURL Status: " + resp.Status)
			return false
		}

	}
}

func createKeyValuePairs(m map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String()
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func NotifyPushover(Config string, AppToken string, UserToken string, nTitle string, nBody string, pThumbnail string, nURL string) {
	// NotifyPushover("apb75jkyb1iegxzp4styr5tgidq3fg","RSS Podcast Downloaded (" + pName + ")","<html><body>" + ytvideo_title + "<br /><br />--------------------------------------------<br /><br />" + ytvideo_description + "</body></html>",ytvideo_thumbnail)

	log.Println("-----		")
	log.Println("-----		START NotifyPushover")
	log.Println("-----		")

	// ~~~~~~~~~~~~~~ Print Data ~~~~~~~~~~~~~~~~

	log.Printf("------------------      START Print Data")
	log.Printf("Config: " + Config)
	log.Printf("AppToken: " + AppToken)
	log.Printf("UserToken: " + UserToken)
	log.Printf("nTitle: " + nTitle)
	log.Printf("nBody: " + nBody)
	log.Printf("pThumbnail: " + pThumbnail)
	log.Printf("nURL: " + nURL)
	log.Printf("------------------      END Print Data")

	// ~~~~~~~~~~ Download Thumbnail ~~~~~~~~~~~~

	savename := ""
	if strings.HasSuffix(pThumbnail, ".jpg") {
		savename = "maxresdefault.jpg"
	}

	if strings.HasSuffix(pThumbnail, ".webp") {
		savename = "maxresdefault.webp"
	}

	if strings.HasSuffix(pThumbnail, ".jpg") == false && strings.HasSuffix(pThumbnail, ".webp") == false {
		savename = "maxresdefault.jpg"
	}

	err := DownloadFile(Config+savename, pThumbnail)
	if err != nil {
		// panic(err)
		log.Printf("------------------      START DownloadFile ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END DownloadFile ERROR")
	}
	fmt.Println("Downloaded: " + pThumbnail)

	// ~~~~~~~~~~~~~~ HTTP Post ~~~~~~~~~~~~~~~~~

	out := exec.Command("curl", "-s", "--form-string", "token="+AppToken, "--form-string", "user="+UserToken, "--form-string", "title="+nTitle, "--form-string", "message="+nBody, "--form-string", "html=1", "-F", "attachment=@"+Config+savename, "https://api.pushover.net/1/messages.json")
	out.Stdout = os.Stdout
	out.Stderr = os.Stderr

	if err := out.Run(); err != nil {
		log.Printf("------------------      START NotifyPushover ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END NotifyPushover ERROR")
	}

	log.Println("-----		END NotifyPushover")
}

func Run_YTDLP(sMediaFolder string, sRSSFolder string, RSSTemplate string, HTTPHost string, Config string, pName string, pChannelID string, pFileFormat string, pDownloadArchive string, pFileQuality string, pChannelThumbnail string, PlaylistItems string, pYouTubeURL string, pPushoverAppToken string, pPushoverUserToken string) {
	log.Println("-----		")
	log.Println("-----		Start Run_YTDLP")
	log.Println("-----		")
	// "yt-dlp -v -o /mnt/pve/NFS_1TB/SCRIPTS/DownloadYouTube-Go/podcasts/%%(id)s.%%(ext)s --write-info-json --no-write-playlist-metafiles --playlist-items 1,2 --restrict-filenames --add-metadata --merge-output-format mp4 --format best --abort-on-error --abort-on-unavailable-fragment --no-overwrites --continue --write-description https://www.youtube.com/playlist?list=PLNJTvO4HBij-As-16otoDkTMhSiQ0cyP_"

	// ~~~~~~~~~~~~~~ Print Data ~~~~~~~~~~~~~~~~
	log.Println("-----		")
	log.Printf("sMediaFolder: " + sMediaFolder)
	log.Printf("sRSSFolder: " + sRSSFolder)
	log.Printf("RSSTemplate: " + RSSTemplate)
	log.Printf("HTTPHost: " + HTTPHost)
	log.Printf("Config: " + Config)
	log.Printf("pName: " + pName)
	log.Printf("pChannelID: " + pChannelID)
	log.Printf("pFileFormat: " + pFileFormat)
	log.Printf("pDownloadArchive: " + pDownloadArchive)
	log.Printf("pFileQuality: " + pFileQuality)
	log.Printf("pChannelThumbnail: " + pChannelThumbnail)
	log.Printf("PlaylistItems: " + PlaylistItems)
	log.Printf("pYouTubeURL: " + pYouTubeURL)
	log.Printf("pPushoverAppToken: " + pPushoverAppToken)
	log.Printf("pPushoverUserToken: " + pPushoverUserToken)
	log.Println("-----		")

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	if pChannelID != "TikTok" {
		// =========================================================
		// ============= Download Channel JSON Only ================
		// =========================================================

		dlname := pChannelID + "/" + pChannelID + ".%(ext)s"

		log.Println("-----		")
		log.Println("-----		Start Download Channel JSON Only")
		log.Println("-----		")

		// out, err := exec.Command("yt-dlp", "-v", "-o", fmt.Sprintf("%s/%s", sMediaFolder, dlname), "--playlist-items", "0", "--write-info-json", "--restrict-filenames", "--add-metadata", "--merge-output-format", pFileFormat, "--format", pFileQuality, "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", pYouTubeURL).Output()

		out := exec.Command("yt-dlp", "-v", "-o", sMediaFolder+dlname, "--playlist-items", "0", "--write-info-json", "--restrict-filenames", "--add-metadata", "--merge-output-format", pFileFormat, "--format", pFileQuality, "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", pYouTubeURL)
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr

		if err := out.Run(); err != nil {
			log.Printf("------------------      START YT-DLP Channel JSON Only ERROR")
			log.Fatal(err.Error())
			log.Printf("------------------      END YT-DLP Channel JSON Only ERROR")

		}
	}
	// =========================================================
	// ============= Download Videos with yt-dlp ===============
	// =========================================================

	dlname2 := pChannelID + "/" + "%(id)s.%(ext)s"

	log.Println("-----		")
	log.Println("-----		Download Videos with yt-dlp")
	log.Println("-----		")

	out2 := exec.Command("yt-dlp", "-v", "-o", sMediaFolder+dlname2, "--playlist-items", PlaylistItems, "--write-info-json", "--no-write-playlist-metafiles", "--download-archive", pDownloadArchive, "--restrict-filenames", "--add-metadata", "--merge-output-format", pFileFormat, "--format", pFileQuality, "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", "--write-description", pYouTubeURL)
	out2.Stdout = os.Stdout
	out2.Stderr = os.Stderr

	if err := out2.Run(); err != nil {
		log.Printf("------------------      START YT-DLP ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END YT-DLP ERROR")

	}

	// =========================================================
	// ================ List Downloaded Files ==================
	// =========================================================

	log.Println("-----		")
	log.Println("-----		List Downloaded Files")
	log.Println("-----		")
	directory := sMediaFolder + pChannelID
	descfiles, descerr := WalkMatch(directory+"/", "*.description")

	if descerr != nil {
		log.Printf("------------------      START List Downloaded Files ERROR")
		// log.Printf("%s", descerr)
		log.Fatal(descerr)
		log.Printf("------------------      END List Downloaded Files ERROR")
	}

	log.Println("-----		")
	log.Println("-----		List Files to add to RSS Feed")
	log.Println("-----		")
	for _, fname := range descfiles {
		// ------- Get Files ---------
		arrfname_noext := strings.Split(fname, ".")
		fname_noext := arrfname_noext[0]
		fname_json := fname_noext + ".info.json"
		fname_mp3 := fname_noext + ".mp3"
		fname_mp4 := fname_noext + ".mp4"
		fname_description := fname_noext + ".description"

		log.Println("fname_noext: " + fname_noext)
		log.Println("fname_mp3: " + fname_mp3)
		log.Println("fname_mp4: " + fname_mp4)
		log.Println("fname_description: " + fname_description)
		log.Println("fname_json: " + fname_json)

		//  Check if Paths are Valid --
		filename_json_isfile := IsValid(fname_json)
		filename_mp3_isfile := IsValid(fname_mp3)
		filename_mp4_isfile := IsValid(fname_mp4)

		if filename_json_isfile == true {
			log.Println("The JSON file is present.")
		}
		if filename_mp3_isfile == true {
			log.Println("The MP3 file is present.")
			// filename_ext := fname_mp3
		}
		if filename_mp4_isfile == true {
			log.Println("The MP4 file is present.")
			// filename_ext := fname_mp4
		}

		log.Println("-----		")
		log.Println("-----		Get JSON Information")
		log.Println("-----		")

		if filename_json_isfile == true && filename_mp4_isfile == true {
			// //  Open and Read JSON file --
			// Let's first read the `config.json` file
			content, contenterr := ioutil.ReadFile(fname_json)
			if contenterr != nil {
				log.Fatal("Error when opening file: ", contenterr)
			}

			// defining a map
			var mapresult map[string]interface{}
			maperr := json.Unmarshal([]byte(content), &mapresult)

			if maperr != nil {
				// print out if error is not nil
				// fmt.Println(maperr)
				log.Fatal("Error reading JSON File ", maperr)
			}

			var jsonpayload JsonData
			jsonpayload.channel_url = ""
			jsonpayload.description = ""
			jsonpayload.duration_string = "0:0"
			jsonpayload.id = ""
			jsonpayload.thumbnail = ""
			jsonpayload.title = ""
			jsonpayload.uploader_url = ""
			jsonpayload.webpage_url = ""
			// jsonpayload.filesize_approx = 0.0

			jsonpayload.id = fmt.Sprint(mapresult["id"])
			jsonpayload.title = fmt.Sprint(mapresult["title"])
			jsonpayload.thumbnail = fmt.Sprint(mapresult["thumbnail"])
			jsonpayload.description = fmt.Sprint(mapresult["description"])
			jsonpayload.uploader_url = fmt.Sprint(mapresult["uploader_url"])
			jsonpayload.channel_url = fmt.Sprint(mapresult["channel_url"])
			jsonpayload.webpage_url = fmt.Sprint(mapresult["webpage_url"])
			jsonpayload.duration_string = fmt.Sprint(mapresult["duration_string"])
			// jsonpayload.filesize_approx = mapresult["filesize_approx"].(float64)
			// var Filesize float64
			// Filesize = (float64(jsonpayload.filesize_approx) / 1024) / 1024
			// jsonpayload.filesize_approx = roundFloat(Filesize, 2)

			// -- Test Thumbnail Path ----
			ytvideo_thumbnail := "https://i.ytimg.com/vi_webp/" + jsonpayload.id + "/maxresdefault.webp"
			ValidURI := IsValidURL(ytvideo_thumbnail)
			if ValidURI == true {
				jsonpayload.thumbnail = ytvideo_thumbnail
			}

			ytvideo_thumbnail2 := "https://i.ytimg.com/vi_webp/" + jsonpayload.id + "/maxresdefault.jpg"
			ValidURI2 := IsValidURL(ytvideo_thumbnail2)
			if ValidURI2 == true {
				jsonpayload.thumbnail = ytvideo_thumbnail2
			}

			// --- Print Final Data ------

			log.Printf("jsonpayload.id: " + jsonpayload.id)
			log.Printf("jsonpayload.title: " + jsonpayload.title)
			log.Printf("jsonpayload.thumbnail: " + jsonpayload.thumbnail)
			// log.Printf("jsonpayload.description: " + jsonpayload.description)
			log.Printf("jsonpayload.uploader_url: " + jsonpayload.uploader_url)
			log.Printf("jsonpayload.channel_url: " + jsonpayload.channel_url)
			log.Printf("jsonpayload.webpage_url: " + jsonpayload.webpage_url)
			log.Printf("jsonpayload.duration_string: " + jsonpayload.duration_string)
			// log.Printf("jsonpayload.filesize_approx: " + fmt.Sprint(jsonpayload.filesize_approx))

			// =========================================================
			// ======== Proceed if RSS XML File Doesn't exist ==========
			// =========================================================
			log.Println("-----		")
			log.Println("-----		Proceed if RSS XML File Doesn't exist")
			log.Println("-----		")
			rssPathFile := sRSSFolder + pChannelID + "RSS.xml"
			log.Printf("rssPathFile: " + rssPathFile)
			rssPathFile_Valid := IsValid(rssPathFile)
			if rssPathFile_Valid == false {
				log.Println("-----		")
				log.Println("-----		Get JSON Channel Information")
				log.Println("-----		")
				// =========================================================
				// =============== Get Channel Information =================
				// =========================================================

				channel_filename_json := sMediaFolder + pChannelID + "/" + pChannelID + ".info.json"
				content2, contenterr2 := ioutil.ReadFile(channel_filename_json)
				if contenterr2 != nil {
					log.Fatal("Error when opening file: ", contenterr2)
				}

				// defining a map
				var mapresult2 map[string]interface{}
				maperr2 := json.Unmarshal([]byte(content2), &mapresult2)

				if maperr2 != nil {
					log.Fatal("Error reading JSON File ", maperr2)
				}

				var jsonchannelpayload JsonChannelData
				jsonchannelpayload.description = ""
				jsonchannelpayload.thumbnail = ""

				// ~~~~~~~~~~~ Get Description ~~~~~~~~~~~~~~
				jsonchannelpayload.description = fmt.Sprint(mapresult2["description"])

				if pChannelThumbnail == "" {
					// ~~~~~~~~ Get Channel Thumbnail ~~~~~~~~~~~
					a, _ := json.Marshal(mapresult2["thumbnails"])
					channelthumbjson := string(a)
					var arrresultthumb []map[string]interface{}
					maperrthumb := json.Unmarshal([]byte(channelthumbjson), &arrresultthumb)
					if maperrthumb != nil {
						// print out if error is not nil
						// fmt.Println(maperr)
						log.Fatal("Error reading JSON File ", maperrthumb)
					}

					for i := len(arrresultthumb) - 1; i >= 0; i-- {
						// log.Printf(fmt.Sprintf(arrresultthumb[i]["id"].(string)))
						thumbid := fmt.Sprintf(arrresultthumb[i]["id"].(string))
						thumburl := fmt.Sprintf(arrresultthumb[i]["url"].(string))

						if thumbid == "avatar_uncropped" {
							jsonchannelpayload.thumbnail = thumburl
							break
						}
					}

					// -- Test Thumbnail Path ----
					ValidChannelURI := IsValidURL(jsonchannelpayload.thumbnail)
					if ValidChannelURI == false {
						jsonchannelpayload.thumbnail = ""
					} else {
						pChannelThumbnail = jsonchannelpayload.thumbnail
					}
				} else {
					// -- Test Thumbnail Path ----
					ValidChannelURI := IsValidURL(pChannelThumbnail)
					if ValidChannelURI == false {
						jsonchannelpayload.thumbnail = ""
					}
				}

				// log.Printf("jsonchannelpayload.description: " + jsonchannelpayload.description)
				log.Printf("pChannelThumbnail: " + pChannelThumbnail)

				// ~~~~~~ End Get Channel Thumbnail ~~~~~~~~~

				// =========================================================
				// =================== Create RSS Feed =====================
				// =========================================================

				log.Println("-----		")
				log.Println("-----		Create RSS Feed")
				log.Println("-----		")

				// ~~~~~~~~ Read RSS Template File ~~~~~~~~~~
				log.Println("-----		Read RSS Template File")
				rssTemplateContent, rssTemplateErr := ioutil.ReadFile(RSSTemplate) // the file is inside the local directory
				if rssTemplateErr != nil {
					log.Fatal(rssTemplateErr)
				}

				// ----- Replace Data --------
				rssTemplateData := string(rssTemplateContent)
				rssTemplateData = strings.ReplaceAll(rssTemplateData, "[CHANNEL_LINK]", jsonpayload.channel_url)
				rssTemplateData = strings.ReplaceAll(rssTemplateData, "[PODCAST_TITLE]", pName)
				rssTemplateData = strings.ReplaceAll(rssTemplateData, "[PODCAST_IMAGE]", pChannelThumbnail)
				rssTemplateData = strings.ReplaceAll(rssTemplateData, "[PODCAST_DESCRIPTION]", jsonchannelpayload.description)

				fmt.Println("rssTemplateData:", rssTemplateData)

				// -- Write New RSS File -----
				if writersserr := os.WriteFile(rssPathFile, []byte(rssTemplateData), 0666); writersserr != nil {
					log.Fatal(writersserr)
				}
			}

			// =========================================================
			// ================ Add Items to RSS File ==================
			// =========================================================
			log.Println("-----		")
			log.Println("-----		Create Item XML for RSS File")
			log.Println("-----		")

			log.Println("-----		Read RSS Template File")
			rssContent, rssErr := ioutil.ReadFile(rssPathFile) // the file is inside the local directory
			if rssErr != nil {
				log.Fatal(rssErr)
			}
			RSSData := string(rssContent)

			if strings.Contains(RSSData, jsonpayload.id) {
				log.Printf("Item (" + jsonpayload.id + ") already in RSS file")
			} else {
				// ------ Get PubDate --------
				log.Printf("Item (" + jsonpayload.id + ") not in RSS file")
				PubDateNow := time.Now()
				PubDate := PubDateNow.Format("02/01/2006 03:04:05 -0700")
				log.Printf("PubDate: " + PubDate)

				// ~~~~~~~~~ Replace & in string ~~~~~~~~~~~~
				jsonpayload_thumbnail_amp := strings.ReplaceAll(jsonpayload.thumbnail, "&", "&amp;")

				// ~~~~~ Replace invalid tiktok data ~~~~~~~~
				jsonpayload.channel_url = pYouTubeURL

				// ----- RSS Item Data -------
				RSSItemsData := "\t\t<item>\n\t\t\t<title><![CDATA[" + jsonpayload.title + "]]></title>\n\t\t\t<description><![CDATA[" + jsonpayload.description + "]]></description>\n\t\t\t<link>" + jsonpayload.webpage_url + "</link>\n\t\t\t<guid isPermaLink=\"false\">" + jsonpayload.webpage_url + "</guid>\n\t\t\t<pubDate>" + PubDate + "</pubDate>\n\t\t\t<podcast:chapters url=\"[ITEM_CHAPTER_URL]\" type=\"application/json\"/>\n\t\t\t<itunes:subtitle><![CDATA[" + jsonpayload.uploader_url + "]]></itunes:subtitle>\n\t\t\t<itunes:summary><![CDATA[" + jsonpayload.uploader_url + "]]></itunes:summary>\n\t\t\t<itunes:author><![CDATA[" + jsonpayload.uploader_url + "]]></itunes:author>\n\t\t\t<author><![CDATA[" + jsonpayload.uploader_url + "]]></author>\n\t\t\t<itunes:image href=\"" + jsonpayload_thumbnail_amp + "\"/>\n\t\t\t<itunes:explicit>No</itunes:explicit>\n\t\t\t<itunes:keywords>youtube</itunes:keywords>\n\t\t\t<enclosure url=\"" + HTTPHost + "podcasts/" + pChannelID + "/" + jsonpayload.id + ".mp4" + "\" type=\"video/mpeg\" length=\"" + jsonpayload.duration_string + "\"/>\n\t\t\t<podcast:person href=\"" + jsonpayload.channel_url + "\" img=\"" + jsonpayload_thumbnail_amp + "\">" + jsonpayload.uploader_url + "</podcast:person>\n\t\t\t<podcast:images srcset=\"" + jsonpayload_thumbnail_amp + " 2000w\"/>\n\t\t\t<itunes:duration>" + jsonpayload.duration_string + "</itunes:duration>\n\t\t</item>\n<!-- INSERT_ITEMS_HERE -->"
				RSSData = strings.ReplaceAll(RSSData, "<!-- INSERT_ITEMS_HERE -->", RSSItemsData)

				// -- Add Data to RSS File -----
				if writersserr := os.WriteFile(rssPathFile, []byte(RSSData), 0666); writersserr != nil {
					log.Fatal(writersserr)
				}
				log.Printf("Item added to RSS file: " + jsonpayload.id)

				// =========================================================
				// =================== Notify Pushover =====================
				// =========================================================

				NotifyPushover(Config, pPushoverAppToken, pPushoverUserToken, "RSS Podcast Downloaded ("+pName+")", "<html><body>"+jsonpayload.title+"<br /><br />--------------------------------------------<br /><br />"+jsonpayload.description+"</body></html>", jsonpayload.thumbnail, jsonpayload.webpage_url)
			}
		}
	}
}

func NotifyYouTube(sMediaFolder string, Config string, pName string, pDownloadArchive string, PlaylistItems string, pYouTubeURL string, pPushoverAppToken string, pPushoverUserToken string) {

	log.Println("-----		")
	log.Println("-----		Start NotifyYouTube")
	log.Println("-----		")

	// =========================================================
	// ============= Download Videos with yt-dlp ===============
	// =========================================================

	dlname2 := "%(id)s.%(ext)s"

	out2 := exec.Command("yt-dlp", "-v", "-o", fmt.Sprintf("%s/%s", sMediaFolder, dlname2), "--skip-download", "--playlist-items", PlaylistItems, "--write-info-json", "--no-write-playlist-metafiles", "--download-archive", pDownloadArchive, "--restrict-filenames", "--add-metadata", "--merge-output-format", "mp4", "--format", "best", "--abort-on-error", "--abort-on-unavailable-fragment", "--no-overwrites", "--continue", "--write-description", pYouTubeURL)
	out2.Stdout = os.Stdout
	out2.Stderr = os.Stderr

	if err := out2.Run(); err != nil {
		log.Printf("------------------      START NotifyYouTube YT-DLP ERROR")
		log.Fatal(err.Error())
		log.Printf("------------------      END NotifyYouTube YT-DLP ERROR")

	}

	// =========================================================
	// ================ List Downloaded Files ==================
	// =========================================================

	log.Println("-----		")
	log.Println("-----		List Downloaded Files")
	log.Println("-----		")
	directory := sMediaFolder
	descfiles, descerr := WalkMatch(directory+"/", "*.description")

	if descerr != nil {
		log.Printf("------------------      START List Downloaded Files ERROR")
		// log.Printf("%s", descerr)
		log.Fatal(descerr)
		log.Printf("------------------      END List Downloaded Files ERROR")
	}

	log.Println("-----		")
	log.Println("-----		List Files to add to RSS Feed")
	log.Println("-----		")
	for _, fname := range descfiles {
		// ------- Get Files ---------
		arrfname_noext := strings.Split(fname, ".")
		fname_noext := arrfname_noext[0]
		fname_json := fname_noext + ".info.json"
		// fname_mp3 := fname_noext + ".mp3"
		// fname_mp4 := fname_noext + ".mp4"
		fname_description := fname_noext + ".description"

		log.Println("fname_noext: " + fname_noext)
		// log.Println("fname_mp3: " + fname_mp3)
		// log.Println("fname_mp4: " + fname_mp4)
		log.Println("fname_description: " + fname_description)
		log.Println("fname_json: " + fname_json)

		//  Check if Paths are Valid --
		filename_json_isfile := IsValid(fname_json)
		// filename_mp3_isfile := IsValid(fname_mp3)
		// filename_mp4_isfile := IsValid(fname_mp4)

		if filename_json_isfile == true {
			log.Println("The JSON file is present.")
		}
		// if filename_mp3_isfile == true {
		// 	log.Println("The MP3 file is present.")
		// 	// filename_ext := fname_mp3
		// }
		// if filename_mp4_isfile == true {
		// 	log.Println("The MP4 file is present.")
		// 	// filename_ext := fname_mp4
		// }

		log.Println("-----		")
		log.Println("-----		Get JSON Information")
		log.Println("-----		")

		if filename_json_isfile == true {
			// //  Open and Read JSON file --
			// Let's first read the `config.json` file
			content, contenterr := ioutil.ReadFile(fname_json)
			if contenterr != nil {
				log.Fatal("Error when opening file: ", contenterr)
			}

			// defining a map
			var mapresult map[string]interface{}
			maperr := json.Unmarshal([]byte(content), &mapresult)

			if maperr != nil {
				// print out if error is not nil
				// fmt.Println(maperr)
				log.Fatal("Error reading JSON File ", maperr)
			}

			var jsonpayload JsonData
			jsonpayload.channel_url = ""
			jsonpayload.description = ""
			jsonpayload.duration_string = "0:0"
			jsonpayload.id = ""
			jsonpayload.thumbnail = ""
			jsonpayload.title = ""
			jsonpayload.uploader_url = ""
			jsonpayload.webpage_url = ""
			// jsonpayload.filesize_approx = 0.0

			jsonpayload.id = fmt.Sprint(mapresult["id"])
			jsonpayload.title = fmt.Sprint(mapresult["title"])
			jsonpayload.thumbnail = fmt.Sprint(mapresult["thumbnail"])
			jsonpayload.description = fmt.Sprint(mapresult["description"])
			jsonpayload.uploader_url = fmt.Sprint(mapresult["uploader_url"])
			jsonpayload.channel_url = fmt.Sprint(mapresult["channel_url"])
			jsonpayload.webpage_url = fmt.Sprint(mapresult["webpage_url"])
			jsonpayload.duration_string = fmt.Sprint(mapresult["duration_string"])
			// jsonpayload.filesize_approx = mapresult["filesize_approx"].(float64)
			// var Filesize float64
			// Filesize = (float64(jsonpayload.filesize_approx) / 1024) / 1024
			// jsonpayload.filesize_approx = roundFloat(Filesize, 2)

			// -- Test Thumbnail Path ----
			ytvideo_thumbnail := "https://i.ytimg.com/vi_webp/" + jsonpayload.id + "/maxresdefault.webp"
			ValidURI := IsValidURL(ytvideo_thumbnail)
			if ValidURI == true {
				jsonpayload.thumbnail = ytvideo_thumbnail
			}

			ytvideo_thumbnail2 := "https://i.ytimg.com/vi_webp/" + jsonpayload.id + "/maxresdefault.jpg"
			ValidURI2 := IsValidURL(ytvideo_thumbnail2)
			if ValidURI2 == true {
				jsonpayload.thumbnail = ytvideo_thumbnail2
			}

			// --- Print Final Data ------

			log.Printf("jsonpayload.id: " + jsonpayload.id)
			log.Printf("jsonpayload.title: " + jsonpayload.title)
			log.Printf("jsonpayload.thumbnail: " + jsonpayload.thumbnail)
			// log.Printf("jsonpayload.description: " + jsonpayload.description)
			log.Printf("jsonpayload.uploader_url: " + jsonpayload.uploader_url)
			log.Printf("jsonpayload.channel_url: " + jsonpayload.channel_url)
			log.Printf("jsonpayload.webpage_url: " + jsonpayload.webpage_url)
			log.Printf("jsonpayload.duration_string: " + jsonpayload.duration_string)
			// log.Printf("jsonpayload.filesize_approx: " + fmt.Sprint(jsonpayload.filesize_approx))

			// Clear Donwloaded Files
			os.Remove(fname_description)
			os.Remove(fname_json)

			// ~~~~~~~~~~~~ Add to Archive ~~~~~~~~~~~~~~

			arch, archerr := os.OpenFile(pDownloadArchive, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 644)

			if archerr != nil {
				log.Printf(archerr.Error())
			}
			defer arch.Close()

			if _, err := arch.WriteString("youtube " + jsonpayload.id + "\n"); err != nil {
				log.Fatal(err)
			}

			// =========================================================
			// =================== Notify Pushover =====================
			// =========================================================

			NotifyPushover(Config, pPushoverAppToken, pPushoverUserToken, "RSS YouTube Video Uploaded ("+pName+")", "<html><body>"+jsonpayload.title+"<br /><br />"+jsonpayload.webpage_url+"<br /><br />--------------------------------------------<br /><br />"+jsonpayload.description+"</body></html>", jsonpayload.thumbnail, jsonpayload.webpage_url)
		}
	}
}

// func Run_RSS_YTDLP() {

// }

func main() {
	// name := "Go Developers"
	// log.Println("Hello World:", name)
	xmlFile, err := os.Open("/config/settings.xml")
	// xmlFile, err := os.Open("settingsLOCAL.xml")
	if err != nil {
		log.Println(err)
	}
	// log.Println("Successfully Opened users.xml")

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our PodcastDownload array
	var settingsXML settings
	var validateXML Validate
	// we unmarshal our byteArray which contains our
	// xmlFiles content into 'users' which we defined above
	xml.Unmarshal(byteValue, &settingsXML)

	log.Println("Email: " + settingsXML.Email)
	log.Println("MediaFolder: " + settingsXML.MediaFolder)
	log.Println("MediaFolderNotify: " + settingsXML.MediaFolderNotify)
	log.Println("RSSFolder: " + settingsXML.RSSFolder)
	log.Println("RSSTemplate: " + settingsXML.RSSTemplate)
	log.Println("PushoverUserToken: " + settingsXML.PushoverUserToken)
	log.Println("HTTPHost: " + settingsXML.HTTPHost)
	log.Println("Config: " + settingsXML.Config)

	// =========================================================
	// ================== Validate Settings ====================
	// =========================================================

	log.Println()
	validateXML.MediaFolder = IsValid(settingsXML.MediaFolder)
	validateXML.MediaFolderNotify = IsValid(settingsXML.MediaFolderNotify)
	validateXML.RSSFolder = IsValid(settingsXML.RSSFolder)
	validateXML.RSSTemplate = IsValid(settingsXML.RSSTemplate)
	validateXML.Config = IsValid(settingsXML.Config)

	if settingsXML.PlaylistItems == "" {
		validateXML.PlaylistItems = false
	} else {
		validateXML.PlaylistItems = true
	}

	if settingsXML.PushoverUserToken == "" {
		validateXML.PushoverUserToken = false
	} else {
		validateXML.PushoverUserToken = true
	}

	if settingsXML.HTTPHost == "" {
		validateXML.HTTPHost = false
	} else {
		validateXML.HTTPHost = true
	}

	// ~~~~~~~~ Print Validation Data ~~~~~~~~~~~

	log.Println("MediaFolder Valid: " + fmt.Sprint(validateXML.MediaFolder))
	log.Println("RSSFolder Valid: " + fmt.Sprint(validateXML.RSSFolder))
	log.Println("Config Valid: " + fmt.Sprint(validateXML.Config))
	log.Println("MediaFolderNotify Valid: " + fmt.Sprint(validateXML.MediaFolderNotify))
	log.Println("PushoverUserToken Valid: " + fmt.Sprint(validateXML.PushoverUserToken))
	log.Println("HTTPHost Valid: " + fmt.Sprint(validateXML.HTTPHost))
	log.Println("PlaylistItems Valid: " + fmt.Sprint(validateXML.PlaylistItems))

	// =========================================================
	// =========================================================
	// =========================================================

	// ########################################################################
	// ######################## Loop PodcastDownload ##########################
	// ########################################################################

	if validateXML.MediaFolder == true && validateXML.RSSFolder == true && validateXML.RSSTemplate == true && validateXML.Config == true && validateXML.MediaFolderNotify == true && validateXML.PushoverUserToken == true && validateXML.HTTPHost == true && validateXML.PlaylistItems == true {
		log.Println("-----		")
		log.Println("-----		Start Validate")
		log.Println("-----		")
		log.Println("Valid - MediaFolder")
		log.Println("Valid - RSSFolder")
		log.Println("Valid - RSSTemplate")
		log.Println("Valid - Config")
		// we iterate through every user within our users array and
		// print out the user Type, their name, and their facebook url
		// as just an example
		for i := 0; i < len(settingsXML.PodcastDownload); i++ {
			if settingsXML.PodcastDownload[i].Name == "" && settingsXML.PodcastDownload[i].ChannelID == "" && settingsXML.PodcastDownload[i].ChannelThumbnail == "" && settingsXML.PodcastDownload[i].DownloadArchive == "" && settingsXML.PodcastDownload[i].FileFormat == "" && settingsXML.PodcastDownload[i].FileQuality == "" && settingsXML.PlaylistItems == "" && settingsXML.PodcastDownload[i].YouTubeURL == "" && settingsXML.PodcastDownload[i].PushoverAppToken == "" {
				validateXML.PodcastDownload_Name = false
				validateXML.PodcastDownload_ChannelID = false
				validateXML.PodcastDownload_DownloadArchive = false
				validateXML.PodcastDownload_FileFormat = false
				validateXML.PodcastDownload_FileQuality = false
				validateXML.PodcastDownload_YouTubeURL = false
				validateXML.PlaylistItems = false
				log.Println("Not Valid - PodcastDownload_Name")
				log.Println("Not Valid - PodcastDownload_ChannelID")
				log.Println("Not Valid - PodcastDownload_DownloadArchive")
				log.Println("ot Valid - PodcastDownload_FileFormat")
				log.Println("Not Valid - PodcastDownload_FileQuality")
				log.Println("Not Valid - PodcastDownload_YouTubeURL")
				log.Println("Not Valid - PlaylistItems")
			} else {
				validateXML.PodcastDownload_DownloadArchive = IsValid(settingsXML.PodcastDownload[i].DownloadArchive)
				if validateXML.PodcastDownload_DownloadArchive == true {
					validateXML.PodcastDownload_Name = true
					validateXML.PodcastDownload_ChannelID = true
					validateXML.PodcastDownload_FileFormat = true
					validateXML.PodcastDownload_FileQuality = true
					validateXML.PodcastDownload_YouTubeURL = true
					validateXML.PlaylistItems = true
					log.Println("Valid - PodcastDownload_Name")
					log.Println("Valid - PodcastDownload_ChannelID")
					log.Println("Valid - PodcastDownload_DownloadArchive")
					log.Println("Valid - PodcastDownload_FileFormat")
					log.Println("Valid - PodcastDownload_FileQuality")
					log.Println("Valid - PodcastDownload_YouTubeURL")
					log.Println("Valid - PlaylistItems")
				} else {
					log.Println("Valid - PodcastDownload_Name")
					log.Println("Valid - PodcastDownload_ChannelID")
					log.Println("Not Valid - PodcastDownload_DownloadArchive")
					log.Println("Valid - PodcastDownload_FileFormat")
					log.Println("Valid - PodcastDownload_FileQuality")
					log.Println("Valid - PodcastDownload_YouTubeURL")
					log.Println("Valid - PlaylistItems")
				}
			}
			log.Println("-----		")
			log.Println("-----		End Validate")
			log.Println("-----		")
			log.Println("")

			// =========================================================
			// =================== Check All Valid =====================
			// =========================================================

			if validateXML.MediaFolder == true && validateXML.PodcastDownload_ChannelID == true && validateXML.PodcastDownload_DownloadArchive == true && validateXML.PodcastDownload_FileFormat == true && validateXML.PodcastDownload_FileQuality == true && validateXML.PodcastDownload_Name == true && validateXML.PlaylistItems == true && validateXML.PodcastDownload_YouTubeURL == true {
				log.Println("-----		")
				log.Println("-----		PodcastDownload")
				log.Println("-----		")
				log.Println("PodcastDownload.Name: " + settingsXML.PodcastDownload[i].Name)
				log.Println("PodcastDownload.ChannelID: " + settingsXML.PodcastDownload[i].ChannelID)
				log.Println("PodcastDownload.ChannelThumbnail: " + settingsXML.PodcastDownload[i].ChannelThumbnail)
				log.Println("PodcastDownload.DownloadArchive: " + settingsXML.PodcastDownload[i].DownloadArchive)
				log.Println("PodcastDownload.FileFormat: " + settingsXML.PodcastDownload[i].FileFormat)
				log.Println("PodcastDownload.FileQuality: " + settingsXML.PodcastDownload[i].FileQuality)
				log.Println("PodcastDownload.YouTubeURL: " + settingsXML.PodcastDownload[i].YouTubeURL)
				log.Println("PlaylistItems: " + settingsXML.PlaylistItems)
				log.Println("-----		")

				Run_YTDLP(settingsXML.MediaFolder, settingsXML.RSSFolder, settingsXML.RSSTemplate, settingsXML.HTTPHost, settingsXML.Config, settingsXML.PodcastDownload[i].Name, settingsXML.PodcastDownload[i].ChannelID, settingsXML.PodcastDownload[i].FileFormat, settingsXML.PodcastDownload[i].DownloadArchive, settingsXML.PodcastDownload[i].FileQuality, settingsXML.PodcastDownload[i].ChannelThumbnail, settingsXML.PlaylistItems, settingsXML.PodcastDownload[i].YouTubeURL, settingsXML.PodcastDownload[i].PushoverAppToken, settingsXML.PushoverUserToken)
				DeleteOldFiles(settingsXML.MediaFolder + settingsXML.PodcastDownload[i].ChannelID + "/")
				log.Println("")
			}
		}
	}
	// =========================================================
	// =========================================================
	// =========================================================

	// ########################################################################
	// ######################## Loop PodcastsNotifty ##########################
	// ########################################################################

	if validateXML.MediaFolder == true && validateXML.RSSFolder == true && validateXML.RSSTemplate == true && validateXML.Config == true && validateXML.MediaFolderNotify == true && validateXML.PushoverUserToken == true && validateXML.HTTPHost == true && validateXML.PlaylistItems == true {
		log.Println("-----		")
		log.Println("-----		Start PodcastsNotifty")
		log.Println("-----		")
		log.Println("Valid - MediaFolder")
		log.Println("Valid - RSSFolder")
		log.Println("Valid - RSSTemplate")
		log.Println("Valid - Config")
		// we iterate through every user within our users array and
		// print out the user Type, their name, and their facebook url
		// as just an example
		for i := 0; i < len(settingsXML.PodcastsNotifty); i++ {
			if settingsXML.PodcastsNotifty[i].Name == "" && settingsXML.PodcastsNotifty[i].YouTubeURL == "" && settingsXML.PodcastsNotifty[i].PushoverAppToken == "" {
				validateXML.PodcastsNotifty_Name = false
				validateXML.PodcastsNotifty_YouTubeURL = false
				validateXML.PlaylistItems = false
				log.Println("Not Valid - PodcastsNotifty_Name")
				log.Println("Not Valid - PodcastsNotifty_YouTubeURL")
				log.Println("Not Valid - PlaylistItems")
			} else {
				validateXML.PodcastsNotifty_Name = true
				validateXML.PodcastsNotifty_YouTubeURL = true
				validateXML.PlaylistItems = true
				log.Println("Valid - PPodcastsNotifty_Name")
				log.Println("Valid - PodcastsNotifty_YouTubeURL")
				log.Println("Valid - PlaylistItems")
			}
			log.Println("-----		")
			log.Println("-----		End Validate")
			log.Println("-----		")
			log.Println("")

			// =========================================================
			// =================== Check All Valid =====================
			// =========================================================

			if validateXML.MediaFolderNotify == true && validateXML.PodcastsNotifty_Name == true && validateXML.PlaylistItems == true && validateXML.PodcastsNotifty_YouTubeURL == true {
				log.Println("-----		")
				log.Println("-----		PodcastsNotifty")
				log.Println("-----		")
				log.Println("PodcastsNotifty.Name: " + settingsXML.PodcastsNotifty[i].Name)
				log.Println("PodcastsNotifty.YouTubeURL: " + settingsXML.PodcastsNotifty[i].YouTubeURL)
				log.Println("PlaylistItems: " + settingsXML.PlaylistItems)
				log.Println("-----		")

				NotifyYouTube(settingsXML.MediaFolderNotify, settingsXML.Config, settingsXML.PodcastsNotifty[i].Name, settingsXML.Config+"youtube-dl-notify.txt", settingsXML.PlaylistItems, settingsXML.PodcastsNotifty[i].YouTubeURL, settingsXML.PodcastsNotifty[i].PushoverAppToken, settingsXML.PushoverUserToken)
				log.Println("")
			}

			// =========================================================
			// =========================================================
			// =========================================================
		}
	}

	// ########################################################################
	// ####################### Run YT-DLP for TikTok ##########################
	// ########################################################################

	if validateXML.MediaFolder == true && validateXML.RSSFolder == true && validateXML.RSSTemplate == true && validateXML.Config == true && validateXML.PushoverUserToken == true && validateXML.HTTPHost == true && validateXML.PlaylistItems == true {
		log.Println("-----		")
		log.Println("-----		Start RSSDownload")
		log.Println("-----		")
		log.Println("Valid - MediaFolder")
		log.Println("Valid - RSSFolder")
		log.Println("Valid - RSSTemplate")
		log.Println("Valid - Config")

		for i := 0; i < len(settingsXML.RSSDownload); i++ {
			if settingsXML.RSSDownload[i].Name == "" && settingsXML.RSSDownload[i].ChannelID == "" && settingsXML.RSSDownload[i].ChannelThumbnail == "" && settingsXML.RSSDownload[i].DownloadArchive == "" && settingsXML.RSSDownload[i].FileFormat == "" && settingsXML.RSSDownload[i].FileQuality == "" && settingsXML.PlaylistItems == "" && settingsXML.RSSDownload[i].TikTokUsername == "" && settingsXML.RSSDownload[i].TikTokFeed == "" && settingsXML.RSSDownload[i].PushoverAppToken == "" {
				validateXML.RSSDownload_Name = false
				validateXML.RSSDownload_ChannelID = false
				validateXML.RSSDownload_DownloadArchive = false
				validateXML.RSSDownload_FileFormat = false
				validateXML.RSSDownload_FileQuality = false
				validateXML.TikTokUsername = false
				validateXML.TikTokFeed = false
				validateXML.PlaylistItems = false
				log.Println("Not Valid - RSSDownload_Name")
				log.Println("Not Valid - RSSDownload_ChannelID")
				log.Println("Not Valid - RSSDownload_DownloadArchive")
				log.Println("ot Valid - RSSDownload_FileFormat")
				log.Println("Not Valid - RSSDownload_FileQuality")
				log.Println("Not Valid - TikTokUsername")
				log.Println("Not Valid - TikTokFeed")
				log.Println("Not Valid - PlaylistItems")
			} else {
				validateXML.RSSDownload_DownloadArchive = IsValid(settingsXML.RSSDownload[i].DownloadArchive)
				if validateXML.RSSDownload_DownloadArchive == true {
					validateXML.RSSDownload_Name = true
					validateXML.RSSDownload_ChannelID = true
					validateXML.RSSDownload_FileFormat = true
					validateXML.RSSDownload_FileQuality = true
					validateXML.TikTokFeed = true
					validateXML.TikTokUsername = true
					validateXML.PlaylistItems = true
					log.Println("Valid - RSSDownload_Name")
					log.Println("Valid - RSSDownload_ChannelID")
					log.Println("Valid - RSSDownload_DownloadArchive")
					log.Println("Valid - RSSDownload_FileFormat")
					log.Println("Valid - RSSDownload_FileQuality")
					log.Println("Valid - TikTokFeed")
					log.Println("Valid - TikTokUsername")
					log.Println("Valid - PlaylistItems")
				} else {
					log.Println("Valid - RSSDownload_Name")
					log.Println("Valid - RSSDownload_ChannelID")
					log.Println("Not Valid - RSSDownload_DownloadArchive")
					log.Println("Valid - RSSDownload_FileFormat")
					log.Println("Valid - RSSDownload_FileQuality")
					log.Println("Valid - TikTokFeed")
					log.Println("Valid - TikTokUsername")
					log.Println("Valid - PlaylistItems")
				}
			}

			validateXML.TikTokFeed = IsValidURL(settingsXML.RSSDownload[i].TikTokFeed + settingsXML.RSSDownload[i].TikTokUsername)

			log.Println("-----		")
			log.Println("-----		End Validate")
			log.Println("-----		")
			log.Println("")

			// =========================================================
			// =================== Check All Valid =====================
			// =========================================================

			if validateXML.MediaFolder == true && validateXML.RSSDownload_Name == true && validateXML.PlaylistItems == true && validateXML.TikTokFeed == true && validateXML.TikTokUsername == true {
				log.Println("-----		")
				log.Println("-----		RSSDownload")
				log.Println("-----		")
				log.Println("RSSDownload.Name: " + settingsXML.RSSDownload[i].Name)
				log.Println("RSSDownload.ChannelID: " + settingsXML.RSSDownload[i].ChannelID)
				log.Println("RSSDownload.ChannelThumbnail: " + settingsXML.RSSDownload[i].ChannelThumbnail)
				log.Println("RSSDownload.DownloadArchive: " + settingsXML.RSSDownload[i].DownloadArchive)
				log.Println("RSSDownload.FileFormat: " + settingsXML.RSSDownload[i].FileFormat)
				log.Println("RSSDownload.FileQuality: " + settingsXML.RSSDownload[i].FileQuality)
				log.Println("RSSDownload.TikTokFeed: " + settingsXML.RSSDownload[i].TikTokFeed)
				log.Println("RSSDownload.TikTokUsername: " + settingsXML.RSSDownload[i].TikTokUsername)
				log.Println("RSSDownload.RSSURL: " + settingsXML.RSSDownload[i].TikTokFeed + settingsXML.RSSDownload[i].TikTokUsername)
				log.Println("PlaylistItems: " + settingsXML.PlaylistItems)
				log.Println("-----		")

				// ~~~~~~~~~ Read TikTok RSS Feed ~~~~~~~~~~~
				err := DownloadFile(settingsXML.Config+"tiktok.json", settingsXML.RSSDownload[i].TikTokFeed+settingsXML.RSSDownload[i].TikTokUsername)
				if err != nil {
					panic(err)
				}
				log.Println("Downloaded: " + settingsXML.Config + "tiktok.json")

				content, contenterr := ioutil.ReadFile(settingsXML.Config + "tiktok.json")
				if contenterr != nil {
					log.Fatal("Error when opening file: ", contenterr)
				}

				// defining a map
				var mapresult map[string]interface{}
				maperr := json.Unmarshal([]byte(content), &mapresult)

				if maperr != nil {
					// print out if error is not nil
					// fmt.Println(maperr)
					log.Fatal("Error reading JSON File ", maperr)
				}

				var jsonpayload TikTok

				jsonpayload.Icon = fmt.Sprint(mapresult["icon"])
				jsonpayload.Title = fmt.Sprint(mapresult["title"])
				// jsonpayload. = fmt.Sprint(mapresult["items"])

				log.Println("icon: " + jsonpayload.Icon)
				log.Println("title: " + jsonpayload.Title)
				log.Println("RSSFolder: " + settingsXML.RSSFolder)
				log.Println("RSSTemplate: " + settingsXML.RSSTemplate)
				log.Println("HTTPHost: " + settingsXML.HTTPHost)
				log.Println("Config: " + settingsXML.Config)

				// ~~~~~~~~~~ Loop through Items ~~~~~~~~~~~~

				var jsonitemspayload Entry
				jsonitemspayload.Link = ""
				jsonitemspayload.Title = ""

				a, _ := json.Marshal(mapresult["items"])
				rssitemjson := string(a)
				var arrresultitem []map[string]interface{}
				maperrthumb := json.Unmarshal([]byte(rssitemjson), &arrresultitem)
				if maperrthumb != nil {
					// print out if error is not nil
					// fmt.Println(maperr)
					log.Fatal("Error reading JSON File ", maperrthumb)
				}

				for j := 0; j < 5; j++ {
					// log.Printf(fmt.Sprintf(arrresultthumb[i]["id"].(string)))
					// thumbid := fmt.Sprintf(arrresultitem[i]["id"].(string))
					// thumburl := fmt.Sprintf(arrresultitem[i]["url"].(string))
					jsonitemspayload.Title = fmt.Sprintf(arrresultitem[j]["title"].(string))
					jsonitemspayload.Link = fmt.Sprintf(arrresultitem[j]["url"].(string))
					log.Println("---  Item " + fmt.Sprint(i) + ": " + settingsXML.Config)
					log.Println("jsonitemspayload.Title: " + jsonitemspayload.Title)
					log.Println("jsonitemspayload.Link: " + jsonitemspayload.Link)

					// Run_YTDLP(settingsXML.MediaFolder, settingsXML.Config, settingsXML.RSSDownload[i].Name, settingsXML.RSSDownload[i].DownloadArchive, settingsXML.PlaylistItems, jsonitemspayload.Link)

					Run_YTDLP(settingsXML.MediaFolder, settingsXML.RSSFolder, settingsXML.RSSTemplate, settingsXML.HTTPHost, settingsXML.Config, settingsXML.RSSDownload[i].Name, settingsXML.RSSDownload[i].ChannelID, settingsXML.RSSDownload[i].FileFormat, settingsXML.RSSDownload[i].DownloadArchive, settingsXML.RSSDownload[i].FileQuality, settingsXML.RSSDownload[i].ChannelThumbnail, settingsXML.PlaylistItems, jsonitemspayload.Link, settingsXML.RSSDownload[i].PushoverAppToken, settingsXML.PushoverUserToken)
					DeleteOldFiles(settingsXML.MediaFolder + settingsXML.RSSDownload[i].ChannelID + "/")
				}

				// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

				log.Println("")
			}

			// =========================================================
			// =========================================================
			// =========================================================
		}
	}

	// ########################################################################
	// ########################################################################
	// ########################################################################

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()
}
