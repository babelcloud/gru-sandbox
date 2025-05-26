import axios, { AxiosInstance } from 'axios';

function getHttp(baseURL: string, apiKey: string): AxiosInstance {
    const http = axios.create({
        baseURL: baseURL,
        headers: {
            'Content-Type': 'application/json',
            "Authorization": `Bearer ${apiKey}`,
        },
    });

    // 添加响应拦截器处理错误
    http.interceptors.response.use(
        response => response,
        error => {
            const errorMessage = error.response
                ? `请求错误: ${error.response.status} ${error.response.statusText} - ${JSON.stringify(error.response.data)}`
                : error.request
                    ? '请求已发送但未收到响应'
                    : `请求配置错误: ${error.message}`;
            
            console.error('HTTP错误:', errorMessage);
            
            return Promise.reject(errorMessage);
        }
    );

    return http;
}

export default getHttp; 