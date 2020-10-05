/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:38:30
 * @Last Modified: thonsun, 2020-10-05  16:38:30
 */

package gateway

import (
	"bytes"
	"html/template"
	"net/http"

	"asec/models"
)

// GenerateBlockPage ...
func GenerateBlockPage(w http.ResponseWriter, hitInfo *models.HitInfo) {
	tmpl := template.New("asec")
	tmpl, _ = tmpl.Parse(blockHTML)
	w.WriteHeader(403)
	tmpl.Execute(w, hitInfo)
}

// GenerateBlockConcent ...
func GenerateBlockConcent(hitInfo *models.HitInfo) []byte {
	tmpl := template.New("asec")
	tmpl, _ = tmpl.Parse(blockHTML)
	buf := new(bytes.Buffer)
	tmpl.Execute(buf, hitInfo)
	return buf.Bytes()
}

var blockHTML = `<!DOCTYPE html>
<html>
<head>
<title>403 Forbidden</title>
</head>
<style>
body {
    font-family: Arial, Helvetica, sans-serif;
    text-align: center;
}

.text-logo {
    display: block;
	width: 260px;
    font-size: 48px;  
    background-color: #F9F9F9;    
    color: #f5f5f5;    
    text-decoration: none;
    text-shadow: 2px 2px 4px #000000;
    box-shadow: 2px 2px 3px #D5D5D5;
    padding: 15px; 
    margin: auto;    
}

.block_div {
    padding: 10px;
    width: 70%;    
    margin: auto;
}

</style>
<body>
<div class="block_div">
<a href="http://www.asec.com/" target="_blank" class="text-logo">asec</a>
<hr>
Reason: {{.VulnName}}, Policy ID: {{.PolicyID}}, by asec Application Gateway
</div>
</body>
</html>
`
