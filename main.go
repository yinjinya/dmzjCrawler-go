package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// {卷名1: [urls], 卷名2: [urls]}
func parseChaptersJSON(comicJSON *ComicJSON) map[string][]ChapterSingleData {
	chapters := comicJSON.Chapters
	volumes := make(map[string][]ChapterSingleData, len(chapters))
	for _, dict := range chapters {
		volumeName := dict.Title
		volumeDetail := dict.Data
		volumes[volumeName] = volumeDetail
	}
	return volumes
}

func saveChapterPages(chapter *Chapter, comicName, volumeName string) bool {
	for i, pageURL := range chapter.PageURLs {
		bytes, ok := getPage(pageURL)
		if !ok {
			return false
		}
		filedirpath := fmt.Sprintf("comic/%s/%s/%s", comicName, volumeName, chapter.Title)
		if err := os.MkdirAll(filedirpath, 0666); err != nil {
			fmt.Fprintf(os.Stderr, "目录创建失败. %v", err)
			return false
		}
		filepath := filedirpath + fmt.Sprintf("/%d.jpg", i)
		if err := writeFileBytes(bytes, filepath); err != nil {
			return false
		}
	}
	return true
}

func cacheJSON(v interface{}, filedirpath, filename string, flagCache bool) bool {
	if flagCache {
		var err error
		content, err := JSONToBytes(v)
		if err != nil {
			return false
		}
		err = os.MkdirAll(filedirpath, 0666)
		if err != nil {
			return false
		}
		err = writeFileBytes(content, filedirpath+"/"+filename+".json")
		if err != nil {
			return false
		}
		return true
	}
	return false
}

var flagSearch = flag.Bool("search", false, "搜索漫画名")
var flagComicID = flag.Int("id", 0, "直接使用漫画ID来下载")
var flagCache = flag.Bool("cache", false, "是否缓存json文件")

func main() {
	flag.Parse()

	var comicID int

	var flagMode byte
	var query string
	if *flagComicID != 0 {
		comicID = *flagComicID
	}else if len(flag.Args()) == 0 {
		// comicID = 38342
		
		fmt.Println("搜索标题or查找漫画ID?(a/b)")
		fmt.Scanf("%c\n", &flagMode)
		if flagMode == 'a' {
			fmt.Scanf("%s", &query)
		} else {
			fmt.Scanf("%d", comicID)
		}
		
	}
	if *flagSearch || flagMode == 'a' {
		query = strings.Join(flag.Args(), " ")
		searchJSON, err := getSearchJSON(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "搜索失败. %v", err)
			os.Exit(1)
			exec.Command("pause")
		}
		cacheJSON(searchJSON, "cache/search", query+".json", *flagCache)
		var chooseNum int
		for i := 0; i < 15 && i < len(searchJSON); i++ {
			searchSingleJSON := searchJSON[i]
			fmt.Printf("%-3d %d\t%s\t%s\n", i, searchSingleJSON.ID, searchSingleJSON.ComicName, searchSingleJSON.ComicURL)
		}
		fmt.Scanf("%d", &chooseNum)
		comicID = searchJSON[chooseNum].ID
	}


	var comicJSON *ComicJSON
	comicJSON, err := getComicJSON(comicID)
	if err != nil {
		os.Exit(1)
	}
	cacheJSON(comicJSON, fmt.Sprintf("cache/comic/%d", comicID), fmt.Sprintf("%d.json", comicID), *flagCache)
	volumes := parseChaptersJSON(comicJSON)
	for volumeName, volumeDetail := range volumes {
		fmt.Println(volumeName)
		for i, chapterDict := range volumeDetail {
			chapterID := chapterDict.ChapterID
			chapterTitle := chapterDict.ChapterTitle

			var chapter *Chapter
			chapter, err := getChapterJSON(comicID, chapterID)
			if err != nil {
				os.Exit(1)
			}
			cacheJSON(chapter, fmt.Sprintf("cache/comic/%d/chapters", comicID), fmt.Sprintf("%d.json", comicID), *flagCache)
			picNum := chapter.PicNum
			fmt.Printf("第%d章\t%s\t%d图片数:%d\n", i+1, chapterTitle, chapterID, picNum)
			ok := saveChapterPages(chapter, comicJSON.ComicName, volumeName)
			if !ok {
				fmt.Fprintf(os.Stderr, "章节图片下载失败.")
				os.Exit(1)
			}
		}
	}
}