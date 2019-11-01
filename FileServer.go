package main
import(
	"fmt"
	"os"
	"net/http"
	"strings"
	"errors"
	"strconv"
	"math/rand"
	"path/filepath"
	"io"
	"net/url"
)

var dirPath=""
func main(){
	args:=os.Args
	if len(args)>1{
		dirPath=strings.TrimSpace(args[1])
		if dirPath==""{
			panic(errors.New("请指定要分享的目录"))
		}
	}

	dir:=isDir(dirPath)
	if !dir{
		panic(errors.New(dirPath+"不是一个目录"))
	}
	dirPath,_=filepath.Abs(dirPath)
	httpFileHandler:=HttpFileHandler{dirPath}
	for ;;{
		port:=rand.Intn(65535)
		fmt.Println(`使用`,port,`端口启动。。。`)
		err:=http.ListenAndServe(":"+strconv.Itoa(port),httpFileHandler)
		fmt.Println("---->",err,"<----")
		if err!=nil{
			fmt.Printf("%T",err)
		}
	}

}

type HttpFileHandler struct{
	RootDir string
}

func(fileHandler HttpFileHandler)ServeHTTP(responseWriter http.ResponseWriter, request *http.Request){
	reqPath:=request.RequestURI
	reqPathParts:=strings.Split(reqPath,`/`)
	reqPathUnescaped:=""
	for i:=0;i<len(reqPathParts)-1;i++{
		unescaped,_:=url.QueryUnescape(reqPathParts[i])
		reqPathUnescaped+=unescaped+`/`
	}
	unescaped,_:=url.QueryUnescape(reqPathParts[len(reqPathParts)-1])
	reqPathUnescaped+=unescaped
	wholePath:=fileHandler.RootDir+reqPathUnescaped
	reqAbsPath,_:=filepath.Abs(wholePath)
	fmt.Println("----->")
	fmt.Println(reqPath)
	fmt.Println(reqPathUnescaped)
	fmt.Println(reqAbsPath)
	fmt.Println("<-----")
	if strings.HasPrefix(reqAbsPath,fileHandler.RootDir){
		if isFile(reqAbsPath){
			fileInfo,_:=os.Stat(reqAbsPath)
			file,openErr:=os.Open(reqAbsPath)
			if openErr==nil{
				buf:=make([]byte,1024*1024*128)
				responseWriter.Header().Set(`Content-Type`,`application/octet-stream`)
				responseWriter.Header().Set(`Content-Length`,strconv.Itoa(int(fileInfo.Size())))
				responseWriter.Header().Set(`Content-Disposition`,`attachment;filename=`+url.QueryEscape(fileInfo.Name()))
				responseWriter.Header().Set(`Content-Transfer-Encoding`,`binary`)
				responseWriter.WriteHeader(200)
				for ;;{
					l,readErr:=file.Read(buf)
					if l>0{
						responseWriter.Write(buf[:l])
					}else{
						break;
					}
					if readErr==io.EOF{
						break;
					}
				}
			}else{
				responseWriter.Header().Set(`Content-Type`,`text/html;charset=UTF-8`)
				responseWriter.WriteHeader(500)
				responseWriter.Write([]byte("该文件无法下载"))
			}

		}else if isDir(reqAbsPath){
			dir,err:=os.Open(reqAbsPath)
			// relative:=strings.TrimPrefix(reqAbsPath,fileHandler.RootDir)
			// relative=strings.ReplaceAll(relative,`\`,`/`)
			// href:=""
			// parts:=strings.Split(relative,`/`)
			// for i:=0;i<len(parts)-1;i++{
			// 	href+=url.QueryEscape(parts[i])+`/`
			// }
			// href+=url.QueryEscape(parts[len(parts)-1])
			if err==nil{
				names,err1:=dir.Readdirnames(-1)
				if err1==nil{
					content:="<table>"
					for _,v:=range names{
						content+=`<tr><td><a href="`+url.QueryEscape(v)+`">`+v+`</a></td></tr>`
					}
					content+="</table>"
					responseWriter.Header().Set(`Content-Type`,`text/html;charset=UTF-8`)
					responseWriter.WriteHeader(200)
					responseWriter.Write([]byte(content))
				}else{
					responseWriter.WriteHeader(500)
					responseWriter.Header().Set(`Content-Type`,`text/html;charset=UTF-8`)
					responseWriter.Write([]byte("服务器发生错误"))
				}
			}else{
				responseWriter.Header().Set(`Content-Type`,`text/html;charset=UTF-8`)
				responseWriter.WriteHeader(500)
				responseWriter.Write([]byte("服务器发生错误"))
			}
		}else{
			responseWriter.Header().Set(`Content-Type`,`text/html;charset=UTF-8`)
			responseWriter.WriteHeader(500)
			responseWriter.Write([]byte("服务器发生错误"))
		}
	}else{
		responseWriter.Header().Set(`Content-Type`,`text/html;charset=UTF-8`)
		responseWriter.WriteHeader(404)
		responseWriter.Write([]byte("资源找不到"))
	}
}

func isFile(name string)bool{
	if fileInfo,err:=os.Stat(name);err!=nil{
		return false
	}else{
		return fileInfo.Mode().IsRegular()
	}
}

func isDir(name string)bool{
	if fileInfo,err:=os.Stat(name);err!=nil{
		return false
	}else{
		return fileInfo.Mode().IsDir()
	}
}