package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tidusant/c3m/common/c3mcommon"
	"github.com/tidusant/c3m/repo/models"
	"io/ioutil"
	"os"
)

func SubmitTest(c *gin.Context) string {
	content := c.PostForm("data")
	sex, _ := c.Cookie("_s")
	templatename := c.Param("params")

	//build content for test
	buildFolder := rootFolder + "/" + templatename + "/build"
	os.RemoveAll(buildFolder)
	err := os.Mkdir(buildFolder, 0775)
	if err != nil {
		return err.Error()
	}
	//copy all css file
	input, err := ioutil.ReadFile(schemeFolder + "/tailwind.css")
	if err != nil {
		return err.Error()
	}
	err = ioutil.WriteFile(buildFolder+"/tailwind.css", input, 0644)
	if err != nil {
		return err.Error()
	}
	//loop css folder and copy
	if _, err := os.Stat(buildFolder + "/../css"); !os.IsNotExist(err) {
		items, _ := ioutil.ReadDir(buildFolder + "/../css")
		for _, item := range items {
			if !item.IsDir() {
				input, err := ioutil.ReadFile(buildFolder + "/../css/" + item.Name())
				if err != nil {
					return err.Error()
				}
				err = ioutil.WriteFile(buildFolder+"/"+item.Name(), input, 0644)
				if err != nil {
					return err.Error()
				}
			}
		}
	}
	//copy all js file
	if _, err := os.Stat(buildFolder + "/../js"); !os.IsNotExist(err) {
		items, _ := ioutil.ReadDir(buildFolder + "/../js")
		for _, item := range items {
			if !item.IsDir() {
				input, err := ioutil.ReadFile(buildFolder + "/../js/" + item.Name())
				if err != nil {
					return err.Error()
				}
				err = ioutil.WriteFile(buildFolder+"/"+item.Name(), input, 0644)
				if err != nil {
					return err.Error()
				}
			}
		}
	}
	//create content
	err = ioutil.WriteFile(buildFolder+"/content.html", []byte(content), 0644)
	if err != nil {
		return err.Error()
	}

	//call test server to purgecss and minify
	bodystr := c3mcommon.RequestAPI2(lptemplatetestserver+"/purge", templatename, sex)
	return bodystr
	var rs models.RequestResult
	err = json.Unmarshal([]byte(bodystr), &rs)

	if err != nil {
		return err.Error()
	}
	if rs.Status != 1 {

		return rs.Error
	}
	return "ok"
}
