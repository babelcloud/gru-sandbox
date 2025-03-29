package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewBoxInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "inspect",
		Short:              "Get detailed information about a box",
		Long:               "Get detailed information about a box by its ID",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string

			// 解析参数
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxInspectHelp()
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
				default:
					if !strings.HasPrefix(args[i], "-") && boxID == "" {
						boxID = args[i]
					} else if strings.HasPrefix(args[i], "-") {
						fmt.Printf("错误: 未知选项 %s\n", args[i])
						os.Exit(1)
					} else {
						fmt.Printf("错误: 意外的参数 %s\n", args[i])
						os.Exit(1)
					}
				}
			}

			// 验证盒子ID
			if boxID == "" {
				fmt.Println("错误: 需要盒子ID")
				os.Exit(1)
			}

			// 调用API获取盒子详情
			apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s", boxID)
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = fmt.Sprintf("%s/api/v1/boxes/%s", envURL, boxID)
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
				if outputFormat == "json" {
					// JSON格式直接输出
					fmt.Println(string(body))
				} else {
					// 文本格式输出
					fmt.Println("盒子详情:")
					fmt.Println("------------")

					// 解析JSON并格式化输出
					var data map[string]interface{}
					if err := json.Unmarshal(body, &data); err != nil {
						fmt.Printf("错误: 解析JSON响应失败: %v\n", err)
						os.Exit(1)
					}

					// 输出每个键值对
					for key, value := range data {
						// 处理复杂类型
						var valueStr string
						switch v := value.(type) {
						case string, float64, bool, int:
							valueStr = fmt.Sprintf("%v", v)
						default:
							// 对于对象或数组，使用JSON格式
							jsonBytes, err := json.Marshal(v)
							if err != nil {
								valueStr = fmt.Sprintf("%v", v)
							} else {
								valueStr = string(jsonBytes)
							}
						}
						fmt.Printf("%-15s: %s\n", key, valueStr)
					}
				}
			case 404:
				fmt.Printf("盒子未找到: %s\n", boxID)
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			default:
				fmt.Printf("错误: 获取盒子详情失败 (HTTP %d)\n", resp.StatusCode)
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

func printBoxInspectHelp() {
	fmt.Println("用法: gbox box inspect <id> [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("    --output          输出格式 (json或text, 默认: text)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("    gbox box inspect 550e8400-e29b-41d4-a716-446655440000              # 获取盒子详情")
	fmt.Println("    gbox box inspect 550e8400-e29b-41d4-a716-446655440000 --output json  # 获取JSON格式的盒子详情")
	fmt.Println()
}
