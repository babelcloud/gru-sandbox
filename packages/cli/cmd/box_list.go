package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type BoxResponse struct {
	Boxes []struct {
		ID     string `json:"id"`
		Image  string `json:"image"`
		Status string `json:"status"`
	} `json:"boxes"`
}

func NewBoxListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "list",
		Short:              "List all available boxes",
		Long:               "List all available boxes with various filtering options",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string
			var filters []string

			// 解析参数，支持与bash脚本相同的参数格式
			for i := 0; i < len(args); i++ {
				switch args[i] {
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
				case "-f", "--filter":
					if i+1 < len(args) {
						filter := args[i+1]
						if !strings.Contains(filter, "=") {
							fmt.Println("错误: 无效的过滤器格式。使用 field=value")
							os.Exit(1)
						}
						filters = append(filters, filter)
						i++
					} else {
						fmt.Println("错误: --filter 需要参数值")
						os.Exit(1)
					}
				case "--help":
					printBoxListHelp()
					return
				default:
					fmt.Printf("错误: 未知选项 %s\n", args[i])
					os.Exit(1)
				}
			}

			// 如果输出格式未指定，默认为text
			if outputFormat == "" {
				outputFormat = "text"
			}

			// 构建查询参数
			queryParams := ""
			if len(filters) > 0 {
				for i, f := range filters {
					parts := strings.SplitN(f, "=", 2)
					if len(parts) == 2 {
						field := parts[0]
						value := url.QueryEscape(parts[1])

						if i == 0 {
							queryParams = "?filter=" + field + "=" + value
						} else {
							queryParams += "&filter=" + field + "=" + value
						}
					}
				}
			}

			// 调用API服务器
			apiURL := "http://localhost:28080/api/v1/boxes" + queryParams
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = envURL + "/api/v1/boxes" + queryParams
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "请求地址: %s\n", apiURL)
			}

			resp, err := http.Get(apiURL)
			if err != nil {
				fmt.Printf("错误: 调用API失败: %v\n", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("错误: 读取响应失败: %v\n", err)
				os.Exit(1)
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "响应状态码: %d\n", resp.StatusCode)
				fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(body))
			}

			// 处理HTTP状态码
			switch resp.StatusCode {
			case 200:
				var response BoxResponse
				if err := json.Unmarshal(body, &response); err != nil {
					fmt.Printf("错误: 解析JSON响应失败: %v\n", err)
					os.Exit(1)
				}

				if outputFormat == "json" {
					// JSON格式输出
					fmt.Println(string(body))
				} else {
					// 文本格式输出
					if len(response.Boxes) == 0 {
						fmt.Println("未找到盒子")
						return
					}

					// 打印表头
					fmt.Println("ID                                      IMAGE               STATUS")
					fmt.Println("---------------------------------------- ------------------- ---------------")

					// 打印每个盒子的信息
					for _, box := range response.Boxes {
						image := box.Image
						if strings.HasPrefix(image, "sha256:") {
							image = strings.TrimPrefix(image, "sha256:")
							if len(image) > 12 {
								image = image[:12]
							}
						}
						fmt.Printf("%-40s %-20s %s\n", box.ID, image, box.Status)
					}
				}
			case 404:
				fmt.Println("未找到盒子")
			default:
				fmt.Printf("错误: 获取盒子列表失败 (HTTP %d)\n", resp.StatusCode)
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "响应: %s\n", string(body))
				}
				os.Exit(1)
			}
		},
	}

	return cmd
}

func printBoxListHelp() {
	fmt.Println("用法: gbox box list [选项]")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("    gbox box list                              # 列出所有盒子")
	fmt.Println("    gbox box list --output json                # 以JSON格式列出盒子")
	fmt.Println("    gbox box list -f 'label=project=myapp'     # 列出带有project=myapp标签的盒子")
	fmt.Println("    gbox box list -f 'ancestor=ubuntu:latest'  # 列出使用ubuntu:latest镜像的盒子")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("    --output          输出格式 (json或text, 默认: text)")
	fmt.Println("    -f, --filter      过滤盒子 (格式: field=value)")
	fmt.Println("                      支持的字段: id, label, ancestor")
	fmt.Println("                      示例:")
	fmt.Println("                      -f 'id=abc123'")
	fmt.Println("                      -f 'label=project=myapp'")
	fmt.Println("                      -f 'ancestor=ubuntu:latest'")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("    --help            显示帮助信息")
}
