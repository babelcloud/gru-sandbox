package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewBoxStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "stop",
		Short:              "Stop a running box",
		Long:               "Stop a running box by its ID",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string

			// 解析参数
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxStopHelp()
					return
				case "--output":
					if i+1 < len(args) {
						outputFormat = args[i+1]
						if outputFormat != "json" && outputFormat != "text" {
							fmt.Println("错误: 无效的输出格式。必须是 'json' 或 'text'")
							if os.Getenv("TESTING") != "true" {
								os.Exit(1)
							}
							return
						}
						i++
					} else {
						fmt.Println("错误: --output 需要参数值")
						if os.Getenv("TESTING") != "true" {
							os.Exit(1)
						}
						return
					}
				default:
					if !strings.HasPrefix(args[i], "-") && boxID == "" {
						boxID = args[i]
					} else if strings.HasPrefix(args[i], "-") {
						fmt.Printf("错误: 未知选项 %s\n", args[i])
						if os.Getenv("TESTING") != "true" {
							os.Exit(1)
						}
						return
					} else {
						fmt.Printf("错误: 意外的参数 %s\n", args[i])
						if os.Getenv("TESTING") != "true" {
							os.Exit(1)
						}
						return
					}
				}
			}

			// 验证盒子ID
			if boxID == "" {
				fmt.Println("错误: 需要盒子ID")
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}

			// 调用API停止盒子
			apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s/stop", boxID)
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = fmt.Sprintf("%s/api/v1/boxes/%s/stop", envURL, boxID)
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "请求地址: %s\n", apiURL)
			}

			// 创建POST请求
			req, err := http.NewRequest("POST", apiURL, nil)
			if err != nil {
				fmt.Printf("错误: 创建请求失败: %v\n", err)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}
			req.Header.Set("Content-Type", "application/json")

			// 发送请求
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("错误: 调用API失败: %v\n", err)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("错误: 读取响应失败: %v\n", err)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}

			if os.Getenv("DEBUG") == "true" {
				fmt.Fprintf(os.Stderr, "响应状态码: %d\n", resp.StatusCode)
				fmt.Fprintf(os.Stderr, "响应内容: %s\n", string(body))
			}

			// 处理HTTP状态码
			switch resp.StatusCode {
			case 200:
				if outputFormat == "json" {
					fmt.Println(`{"status":"success","message":"盒子已成功停止"}`)
				} else {
					fmt.Println("盒子已成功停止")
				}
			case 404:
				fmt.Printf("盒子未找到: %s\n", boxID)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			default:
				fmt.Printf("错误: 停止盒子失败 (HTTP %d)\n", resp.StatusCode)
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "响应: %s\n", string(body))
				}
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			}
		},
	}

	return cmd
}

func printBoxStopHelp() {
	fmt.Println("用法: gbox box stop <id> [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("    --output          输出格式 (json或text, 默认: text)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("    gbox box stop 550e8400-e29b-41d4-a716-446655440000              # 停止一个盒子")
	fmt.Println("    gbox box stop 550e8400-e29b-41d4-a716-446655440000 --output json  # 停止盒子并输出JSON")
	fmt.Println()
}
