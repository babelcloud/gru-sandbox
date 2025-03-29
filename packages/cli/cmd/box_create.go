package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type BoxCreateRequest struct {
	Image      string            `json:"image,omitempty"`
	Cmd        string            `json:"cmd,omitempty"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	WorkingDir string            `json:"workingDir,omitempty"`
}

type BoxCreateResponse struct {
	ID string `json:"id"`
}

func NewBoxCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "create",
		Short:              "Create a new box",
		Long:               "Create a new box with various options for image, environment, and commands",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string
			var image string
			var command string
			var commandArgs []string
			var env []string
			var labels []string
			var workingDir string

			// 解析参数
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxCreateHelp()
					return
				case "--output":
					if i+1 < len(args) {
						outputFormat = args[i+1]
						if outputFormat != "json" && outputFormat != "text" {
							fmt.Println("错误: 无效的输出格式。必须是 'json' 或 'text'")
							os.Exit(1)
						}
						i++
					} else {
						fmt.Println("错误: --output 需要参数值")
						os.Exit(1)
					}
				case "--image":
					if i+1 < len(args) {
						image = args[i+1]
						i++
					} else {
						fmt.Println("错误: --image 需要参数值")
						os.Exit(1)
					}
				case "--env":
					if i+1 < len(args) {
						envValue := args[i+1]
						if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.+$`).MatchString(envValue) {
							fmt.Println("错误: 无效的环境变量格式。使用 KEY=VALUE")
							os.Exit(1)
						}
						env = append(env, envValue)
						i++
					} else {
						fmt.Println("错误: --env 需要参数值")
						os.Exit(1)
					}
				case "-l", "--label":
					if i+1 < len(args) {
						labelValue := args[i+1]
						if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.+$`).MatchString(labelValue) {
							fmt.Println("错误: 无效的标签格式。使用 KEY=VALUE")
							os.Exit(1)
						}
						labels = append(labels, labelValue)
						i++
					} else {
						fmt.Println("错误: --label 需要参数值")
						os.Exit(1)
					}
				case "-w", "--work-dir":
					if i+1 < len(args) {
						workingDir = args[i+1]
						i++
					} else {
						fmt.Println("错误: --work-dir 需要参数值")
						os.Exit(1)
					}
				case "--":
					if i+1 < len(args) {
						command = args[i+1]
						if i+2 < len(args) {
							commandArgs = args[i+2:]
						}
						i = len(args) // 终止循环
					} else {
						fmt.Println("错误: -- 后需要命令")
						os.Exit(1)
					}
				default:
					fmt.Printf("错误: 未知选项 %s\n", args[i])
					os.Exit(1)
				}
			}

			// 如果输出格式未指定，默认为text
			if outputFormat == "" {
				outputFormat = "text"
			}

			// 构建请求体
			request := BoxCreateRequest{}

			if image != "" {
				request.Image = image
			}

			if command != "" {
				request.Cmd = command
			}

			if len(commandArgs) > 0 {
				request.Args = commandArgs
			}

			if workingDir != "" {
				request.WorkingDir = workingDir
			}

			// 处理环境变量
			if len(env) > 0 {
				request.Env = make(map[string]string)
				for _, e := range env {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						request.Env[parts[0]] = parts[1]
					}
				}
			}

			// 处理标签
			if len(labels) > 0 {
				request.Labels = make(map[string]string)
				for _, l := range labels {
					parts := strings.SplitN(l, "=", 2)
					if len(parts) == 2 {
						request.Labels[parts[0]] = parts[1]
					}
				}
			}

			// 将请求转换为JSON
			requestBody, err := json.Marshal(request)
			if err != nil {
				fmt.Printf("错误: 无法序列化请求: %v\n", err)
				os.Exit(1)
			}

			// 调试输出
			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "请求体:\n")
				var prettyJSON bytes.Buffer
				json.Indent(&prettyJSON, requestBody, "", "  ")
				fmt.Fprintln(os.Stderr, prettyJSON.String())
			}

			// 调用API服务器
			apiURL := "http://localhost:28080/api/v1/boxes"
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = envURL + "/api/v1/boxes"
			}
			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				fmt.Printf("错误: 无法连接到API服务器: %v\n", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			// 读取响应
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("错误: 读取响应失败: %v\n", err)
				os.Exit(1)
			}

			// 检查HTTP状态码
			if resp.StatusCode != 201 {
				fmt.Printf("错误: API服务器返回HTTP %d\n", resp.StatusCode)
				if len(responseBody) > 0 {
					fmt.Printf("响应: %s\n", string(responseBody))
				}
				os.Exit(1)
			}

			// 处理响应
			if outputFormat == "json" {
				// 原样输出JSON
				fmt.Println(string(responseBody))
			} else {
				// 格式化输出
				var response BoxCreateResponse
				if err := json.Unmarshal(responseBody, &response); err != nil {
					fmt.Printf("错误: 解析响应失败: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("已创建盒子，ID为 \"%s\"\n", response.ID)
			}
		},
	}

	return cmd
}

func printBoxCreateHelp() {
	fmt.Println("用法: gbox box create [选项] [--] <命令> [参数...]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("    --output          输出格式 (json或text, 默认: text)")
	fmt.Println("    --image           容器镜像")
	fmt.Println("    --env             环境变量，格式为KEY=VALUE")
	fmt.Println("    -w, --work-dir    工作目录")
	fmt.Println("    -l, --label       自定义标签，格式为KEY=VALUE")
	fmt.Println("    --                命令及其参数")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("    gbox box create --image python:3.9 -- python3 -c 'print(\"Hello\")'")
	fmt.Println("    gbox box create --env PATH=/usr/local/bin:/usr/bin:/bin -w /app -- node server.js")
	fmt.Println("    gbox box create --label project=myapp --label env=prod -- python3 server.py")
	fmt.Println()
}
