package main

import (
	"encoding/json"
	"fmt"
	"github.com/BrokenMyth/go-utils/fileutil"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 设置路由规则
	http.HandleFunc("/file/download", fileDownloadHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/file/create", fileCreateHandler)
	http.HandleFunc("/file/delete", fileDeleteHandler)
	fmt.Println("启动服务...")
	// 启动服务器
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
}

func fileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 获取参数中的文件名
	fileName := r.URL.Query().Get("filename")

	// 读取文件内容
	data, err := ioutil.ReadFile(filepath.Join("res", fileName))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	_ = fileutil.CreateNotExistDir("./res")
	// 读取目录内容
	files, err := ioutil.ReadDir("./res")
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	// 构建 HTML 响应
	var html strings.Builder
	html.WriteString("<h1>Files and Directories</h1><ul>")
	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			name = name + "/"
		}
		html.WriteString(fmt.Sprintf(`<li><a href="/file/download?filename=%s">%s</a> <button onclick="deleteFile('%s')">删除</button></li>`, name, name, name))
	}
	html.WriteString("</ul>")

	// 添加创建文件的按钮
	html.WriteString(`<button onclick="createFile()">创建一个100M的文件</button>`)
	html.WriteString(`
	<script>
		function createFile() {
			fetch('/file/create', {
				method: 'POST',
				body: JSON.stringify({ size: 100 }),
				headers: {
					'Content-Type': 'application/json'
				}
			})
			.then(response => {
				if (response.ok) {
					window.location.reload();
				} else {
					alert('创建文件失败');
				}
			})
			.catch(error => {
				alert('创建文件失败：' + error);
			});
		}
		
		function deleteFile(name) {
			if (confirm("确定要删除文件 " + name + " 吗？")) {
				fetch('/file/delete', {
					method: 'POST',
					body: JSON.stringify({filename: name}),
					headers: {
						'Content-Type': 'application/json'
					}
				})
				.then(response => {
					if (response.ok) {
						window.location.reload();
					} else {
						alert('删除文件失败');
					}
				})
				.catch(error => {
					alert('删除文件失败：' + error);
				});
			}
		}
	</script>`)

	// 设置响应内容类型为 HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html.String()))
}

func fileCreateHandler(w http.ResponseWriter, r *http.Request) {
	// 读取请求体中的 JSON 数据
	type requestBody struct {
		Size int `json:"size"`
	}
	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 生成随机数据并写入文件
	fileName := fmt.Sprintf("%dM", body.Size)
	filePath := filepath.Join("./res", fileName)
	_, err = os.Stat(filePath)
	if err == nil {
		http.Error(w, "File already exists", http.StatusConflict)
		return
	}
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	data := make([]byte, 1024*1024)
	for i := 0; i < body.Size; i++ {
		_, err := rand.Read(data)
		if err != nil {
			http.Error(w, "Failed to generate random data", http.StatusInternalServerError)
			return
		}
		_, err = file.Write(data)
		if err != nil {
			http.Error(w, "Failed to write data to file", http.StatusInternalServerError)
			return
		}
	}

	// 返回成功响应并重定向到当前页面
	url := r.Header.Get("Referer")
	if url == "" {
		url = "/"
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

type deleteFileRequest struct {
	FileName string `json:"filename"`
}

func fileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req deleteFileRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 删除文件
	filePath := filepath.Join("./res", req.FileName)
	err = os.Remove(filePath)
	if err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
}
