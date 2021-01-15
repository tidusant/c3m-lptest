package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidusant/c3m/common/c3mcommon"
	"github.com/tidusant/c3m/repo/models"
	lpmodels "github.com/tidusant/c3mlp/repo/models"
	log "github.com/tidusant/chadmin-log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Template struct {
	Name string
	Path string
}
type Tool struct {
	Name    string
	Title   string
	Icon    string
	Content string
	Child   []Tool
}
type Nav struct {
	Id   string
	Name string
}

func GetTest(c *gin.Context) string {

	name := c.Param("params")
	if name == "" {
		return "Template name empty. "

	}
	//check template exist
	tplFolder := rootFolder + "/" + name
	if _, err := os.Stat(tplFolder); os.IsNotExist(err) {
		return "Template not found. "

	}

	//Get tool
	tools, err := ReadTemplateTool(name)
	if err != nil {
		return err.Error()

	}
	//map tools name to replace in layout content
	mtool := make(map[string]string)
	toolcontent := ""
	trashel := `
<div class="landingpage-trash absolute top-0 hidden bg-opacity-0 z-30" onclick="RemoveItem(this)">
	<div class="bg-black text-white text-xs rounded py-2 px-4 mb-1 right-0 bottom-full">
      Remove item %s {{trashtitle}}     
    </div>
</div>`
	for _, v := range tools {
		if len(v.Child) > 0 {
			toolcontent += `
<div class="cus-not-draggable cursor-pointer hoverable hover:text-white py-2 landingpage-tool-` + v.Name + `">
                        <div class="landingpage-tool-icon">
                          <img class="m-auto" src="` + v.Icon + `" title="` + v.Title + `" />
                        </div>

                        <div class="-mt-8 mega-menu sm:mb-0 shadow-xl bg-white">
`
			for _, v2 := range v.Child {
				toolkey := v.Name + "." + v2.Name
				mtool[toolkey] = v2.Content + fmt.Sprintf(trashel, v2.Title)
				toolcontent += `
<div class="m-auto cursor-pointer float-left p-2 w-max relative" lp-data-id="` + toolkey + `">
                              <div class="landingpage-tool-icon">
                                <img class="m-auto" src="` + v2.Icon + `" title="` + v2.Title + `" />
                              </div>
                            </div>
`
			}
			toolcontent += `<div class="clear-both"></div>
                        </div>
                      </div>`
		} else {

			mtool[v.Name] = v.Content + fmt.Sprintf(trashel, v.Title)
			toolcontent += `
<div class="m-auto py-2 cursor-pointer relative" lp-data-id="` + v.Name + `">
	<div class="landingpage-tool-icon">
	  <img class="m-auto" src="` + v.Icon + `" title="` + v.Title + `" />
	</div>
  </div>
`
		}
	}
	toolcontent = strings.Replace(toolcontent, "{{template_path}}", rootPath+"/"+name, -1)

	//get  layout content
	dat, err := ioutil.ReadFile(tplFolder + "/content.html")
	if err != nil {
		return err.Error()

	}
	var navitems []Nav

	var re = regexp.MustCompile(`\{\{(.*)\}\}`)
	content := string(dat)
	t := re.FindAllStringSubmatch(content, -1)

	for _, v := range t {
		log.Debugf("regex %+v", v[1])
		vtypes := strings.Split(v[1], "_")
		itemname := v[1]

		if len(vtypes) > 1 {
			itemname = vtypes[0]
		}
		if _, ok := mtool[itemname]; ok {
			//parse item type
			itemcontent := mtool[itemname]
			if itemname == "a" {
				itemcontent = strings.Replace(itemcontent, `{{Id}}`, vtypes[1], -1)
				itemcontent = strings.Replace(itemcontent, `{{trashtitle}}`, vtypes[2], -1)
				navitems = append(navitems, Nav{Id: vtypes[1], Name: vtypes[2]})
			} else {
				itemcontent = strings.Replace(itemcontent, `{{trashtitle}}`, "", -1)
			}

			content = strings.Replace(content, `{{`+v[1]+`}}`, `<div class="item-container m-auto landingpage-cursor-pointer relative" lp-data-id="`+v[1]+`">`+itemcontent+`</div>`, -1)
		}
	}

	//======================read and create edit page content============================
	dat, err = ioutil.ReadFile(schemeFolder + "/edit.html")
	if err != nil {
		return err.Error()

	}
	s := string(dat)

	//replace content
	s = strings.Replace(s, "{{toolcontent}}", toolcontent, 1)
	s = strings.Replace(s, "{{pagecontent}}", content, 1)

	b, _ := json.Marshal(mtool)

	s = strings.Replace(s, "{{mtoolcontent}}", string(b), 1)
	s = strings.Replace(s, "{{template_path}}", rootPath+"/"+name, -1)

	//============================nav item============================

	//read nav item template
	dat, err = ioutil.ReadFile(rootFolder + "/" + name + "/navitem.html")
	if err != nil {
		return err.Error()

	}
	navtemplate := string(dat)
	navitemcontent := ``
	log.Debugf("%+v", navitems)
	for _, v := range navitems {
		navitemcontent += strings.Replace(strings.Replace(navtemplate, `{{Name}}`, v.Name, -1), `{{Id}}`, v.Id, -1)
	}
	//replace in content template
	s = strings.Replace(s, "{{navitems}}", navitemcontent, -1)
	s = strings.Replace(s, "{{navitemtemplate}}", navtemplate, 1)

	//============== preview content in iframe
	//render css link
	customcss := ``
	if _, err := os.Stat(tplFolder + "/css"); err == nil {
		files, _ := ioutil.ReadDir(tplFolder + "/css")
		for _, f := range files {
			if !f.IsDir() {
				customcss += `<link href="` + rootPath + "/" + name + `/css/` + f.Name() + `" rel="stylesheet">`
			}
		}
	}
	customcss += `<link href="/scheme/tailwind.css" rel="stylesheet">`
	s = strings.Replace(s, "{{customcss}}", customcss, -1)
	//render js script
	customjs := ``
	if _, err := os.Stat(tplFolder + "/css"); err == nil {
		files, _ := ioutil.ReadDir(tplFolder + "/js")
		for _, f := range files {
			if !f.IsDir() {
				customjs += `<script src="` + rootPath + "/" + name + `/js/` + f.Name() + `"></script>`
			}
		}
	}
	s = strings.Replace(s, "{{customjs}}", customjs, -1)
	s = strings.Replace(s, "{{customiframejs}}", strings.Replace(customjs, `</script>`, `<\/script>`, -1), -1)
	s = strings.Replace(s, "{{templatename}}", name, -1)
	// //Convert your cached html string to byte array
	// c.Writer.Write([]byte(result))
	return s

}

func ReadTemplateTool(tplname string) ([]Tool, error) {
	var tools []Tool
	//read tool item
	path := rootFolder + "/" + tplname
	file, err := os.Open(path + "/items.html")
	if err != nil {
		return tools, err
	}
	defer file.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var tool Tool
	var child Tool
	var contentBuffer bytes.Buffer
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		lineorg := strings.Trim(line, "\n")
		line = RemoveComment(lineorg)
		// Process the line here.
		if strings.Index(line, "#===") == 0 {
			//save content
			if child.Name != "" {
				child.Content = contentBuffer.String()
			} else {
				tool.Content = contentBuffer.String()
			}
			contentBuffer.Reset()
		}

		if line == "#===name===#" {
			//add previous data to tools

			if child.Name != "" {
				tool.Child = append(tool.Child, child)
				//new
				child = Tool{}
			}
			if tool.Name != "" {
				tools = append(tools, tool)
				//new
				tool = Tool{}
			}

			//read next line to get name
			line, err = reader.ReadString('\n')
			if err != nil {
				break
			}
			strs := strings.Split(RemoveComment(line), ":")

			//check name & icon
			if len(strs) < 3 {
				err = fmt.Errorf("name and icon invalid")
				break
			}

			tool.Name = strs[0]
			tool.Title = strs[1]
			tool.Icon = rootPath + "/" + tplname + "/itemicons/" + strs[2]

		} else if line == "#===child===#" {
			if child.Name != "" {
				tool.Child = append(tool.Child, child)
				//new
				child = Tool{}
			}
			//read next line to get name
			line, err = reader.ReadString('\n')
			if err != nil {
				break
			}
			strs := strings.Split(RemoveComment(line), ":")

			//check name & icon
			if len(strs) < 3 {
				err = fmt.Errorf("name and icon invalid")
				break
			}

			child.Name = strs[0]
			child.Title = strs[1]
			child.Icon = rootPath + "/" + tplname + "/itemicons/" + strs[2]
		} else {
			contentBuffer.WriteString(lineorg)
		}
		//===========
	}

	//add last item
	if child.Name != "" {
		child.Content = contentBuffer.String()
		tool.Child = append(tool.Child, child)
	}
	if tool.Name != "" {
		tool.Content = contentBuffer.String()
		tools = append(tools, tool)

	}

	if err != io.EOF {
		return tools, err
	}
	return tools, nil
}
func RemoveComment(s string) string {
	t := strings.Replace(s, `<!--`, ``, 1)
	t2 := strings.Replace(t, `-->`, ``, 1)

	return t2
}
func GetTemplate(session string) ([]lpmodels.Template, error) {
	var rt []lpmodels.Template
	localtemplates := make(map[string]string)

	if _, err := os.Stat(rootFolder); err == nil {
		files, _ := ioutil.ReadDir(rootFolder)
		for _, f := range files {
			if f.IsDir() {
				log.Debugf("check is folder %s", f.Name())
				//check screen shot
				log.Debugf("check %s", rootFolder+"/"+f.Name()+"/screenshot.jpg")
				if _, err := os.Stat(rootFolder + "/" + f.Name() + "/screenshot.jpg"); err == nil {
					localtemplates[f.Name()] = f.Name()
				}
			}
		}
	}

	//get template form server
	bodystr := c3mcommon.RequestAPI(apiserver, "lptpl", session+"|la")
	var rs models.RequestResult
	err := json.Unmarshal([]byte(bodystr), &rs)
	if err != nil {
		return rt, err
	}
	if rs.Status != 1 {
		log.Debugf("rs %+v", rs)
		return rt, fmt.Errorf(rs.Error)
	}

	log.Debugf("rs template: %+v", rs)
	err = json.Unmarshal([]byte(rs.Data), &rt)
	if err != nil {
		return rt, err
	}
	for k, v := range rt {
		//hide templateID & userID
		rt[k].ID = primitive.NilObjectID
		rt[k].UserID = primitive.NilObjectID
		rt[k].Path = "./"
		if _, ok := localtemplates[v.Name]; ok {
			delete(localtemplates, v.Name)
		}
	}

	//add local template into result
	for k, _ := range localtemplates {
		rt = append(rt, lpmodels.Template{Name: k, Status: 0, Path: rootFolder + "/" + k})
	}

	return rt, nil
}
func HandleGetLocal(c *gin.Context) models.RequestResult {
	session := c.PostForm("data")
	templates, err := GetTemplate(session)
	if err != nil {
		errrs := models.RequestResult{Error: err.Error()}
		if err.Error() == "Session not found" {
			errrs.Status = -1
		}
		return errrs
	}
	b, _ := json.Marshal(templates)
	return models.RequestResult{Status: 1, Data: string(b)}
}
