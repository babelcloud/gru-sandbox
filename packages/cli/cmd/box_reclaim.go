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

type BoxReclaimResponse struct {
	Status       string `json:"status"`
	Message      string `json:"message"`
	StoppedCount int    `json:"stoppedCount"`
	DeletedCount int    `json:"deletedCount"`
}

func NewBoxReclaimCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "reclaim",
		Short:              "Reclaim a box resources",
		Long:               "Reclaim a box's resources by force if it's in a stuck state",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string
			var force bool = false

			// 解析参数
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxReclaimHelp()
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
				case "--force", "-f":
					force = true
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

			// 准备API URL
			var apiURL string
			if boxID == "" {
				// 如果没有指定盒子ID，则执行全局回收
				apiURL = "http://localhost:28080/api/v1/boxes/reclaim"
				if envURL := os.Getenv("API_URL"); envURL != "" {
					apiURL = fmt.Sprintf("%s/api/v1/boxes/reclaim", envURL)
				}
			} else {
				// 如果指定了盒子ID，则只回收特定盒子
				apiURL = fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s/reclaim", boxID)
				if envURL := os.Getenv("API_URL"); envURL != "" {
					apiURL = fmt.Sprintf("%s/api/v1/boxes/%s/reclaim", envURL, boxID)
				}
			}

			// 添加强制参数
			if force {
				if strings.Contains(apiURL, "?") {
					apiURL += "&force=true"
				} else {
					apiURL += "?force=true"
				}
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
			req.Header.Set("Accept", "application/json")

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
					// JSON格式直接输出
					fmt.Println(string(body))
				} else {
					// 文本格式输出
					var response BoxReclaimResponse
					if err := json.Unmarshal(body, &response); err != nil {
						fmt.Println("盒子资源已成功回收")
					} else {
						fmt.Println(response.Message)
						if response.StoppedCount > 0 {
							fmt.Printf("已停止 %d 个盒子\n", response.StoppedCount)
						}
						if response.DeletedCount > 0 {
							fmt.Printf("已删除 %d 个盒子\n", response.DeletedCount)
						}
					}
				}
			case 404:
				if boxID != "" {
					fmt.Printf("盒子未找到: %s\n", boxID)
				} else {
					fmt.Println("找不到可回收的盒子")
				}
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			case 400:
				fmt.Printf("错误: 无效的请求: %s\n", string(body))
				if os.Getenv("TESTING") != "true" {
					os.Exit(1)
				}
				return
			default:
				fmt.Printf("错误: 回收盒子资源失败 (HTTP %d)\n", resp.StatusCode)
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

func printBoxReclaimHelp() {
	fmt.Println("用法: gbox box reclaim <id> [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("    --output          输出格式 (json或text, 默认: text)")
	fmt.Println("    -f, --force       强制回收资源，即使盒子正在运行")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("    gbox box reclaim 550e8400-e29b-41d4-a716-446655440000              # 回收盒子资源")
	fmt.Println("    gbox box reclaim 550e8400-e29b-41d4-a716-446655440000 --force      # 强制回收盒子资源")
	fmt.Println("    gbox box reclaim 550e8400-e29b-41d4-a716-446655440000 --output json  # 输出JSON格式结果")
	fmt.Println("    gbox box reclaim                                      # 回收所有符合条件的盒子资源")
	fmt.Println()
}
