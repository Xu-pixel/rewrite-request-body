package traefik_plugin_rewrite_request_body

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strconv"
    "strings"
)

// Config 为插件的可配置项
type Config struct {
    // OldKey 是需要被改名的原始键
    OldKey string `json:"oldKey,omitempty"`
    // NewKey 是新的键名
    NewKey string `json:"newKey,omitempty"`
}

// CreateConfig 返回默认配置
func CreateConfig() *Config {
    return &Config{}
}

// RewriteRequestBody 是中间件实现
type RewriteRequestBody struct {
    next   http.Handler
    name   string
    config *Config
}

// New 创建一个新的中间件实例
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
    return &RewriteRequestBody{
        next:   next,
        name:   name,
        config: config,
    }, nil
}

// ServeHTTP 实现 http.Handler 接口
func (m *RewriteRequestBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    if req == nil || req.Body == nil {
        m.next.ServeHTTP(rw, req)
        return
    }

    // 仅处理 application/json，且未压缩的请求体
    contentType := req.Header.Get("Content-Type")
    if !strings.Contains(strings.ToLower(contentType), "application/json") {
        m.next.ServeHTTP(rw, req)
        return
    }

    if enc := req.Header.Get("Content-Encoding"); enc != "" && strings.ToLower(enc) != "identity" {
        // 压缩或其他编码的请求体不处理
        m.next.ServeHTTP(rw, req)
        return
    }

    oldKey := strings.TrimSpace(m.config.OldKey)
    newKey := strings.TrimSpace(m.config.NewKey)
    if oldKey == "" || newKey == "" || oldKey == newKey {
        m.next.ServeHTTP(rw, req)
        return
    }

    // 读取完整请求体
    originalBody, err := io.ReadAll(req.Body)
    if err != nil {
        // 读取失败则按原样透传
        m.next.ServeHTTP(rw, req)
        return
    }
    // 确保关闭旧 body
    _ = req.Body.Close()

    if len(bytes.TrimSpace(originalBody)) == 0 {
        // 空体直接透传
        req.Body = io.NopCloser(bytes.NewReader(originalBody))
        req.ContentLength = int64(len(originalBody))
        if req.ContentLength >= 0 {
            req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
            req.TransferEncoding = nil
        }
        m.next.ServeHTTP(rw, req)
        return
    }

    // 尝试解析为 JSON 对象（map）或数组
    // 仅在顶层是对象时才执行 key 改名
    var anyJSON interface{}
    if err := json.Unmarshal(originalBody, &anyJSON); err != nil {
        // 非 JSON 或解析失败，原样透传
        req.Body = io.NopCloser(bytes.NewReader(originalBody))
        req.ContentLength = int64(len(originalBody))
        if req.ContentLength >= 0 {
            req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
            req.TransferEncoding = nil
        }
        m.next.ServeHTTP(rw, req)
        return
    }

    modified := false

    if obj, ok := anyJSON.(map[string]interface{}); ok {
        if val, exists := obj[oldKey]; exists {
            obj[newKey] = val
            delete(obj, oldKey)
            modified = true
        }
        // 重新编码
        if modified {
            newBytes, err := json.Marshal(obj)
            if err == nil {
                req.Body = io.NopCloser(bytes.NewReader(newBytes))
                req.ContentLength = int64(len(newBytes))
                if req.ContentLength >= 0 {
                    req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
                    req.TransferEncoding = nil
                }
            } else {
                // 编码失败则回退为原始请求体
                req.Body = io.NopCloser(bytes.NewReader(originalBody))
                req.ContentLength = int64(len(originalBody))
                if req.ContentLength >= 0 {
                    req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
                    req.TransferEncoding = nil
                }
            }
        } else {
            // 未修改则还原 body
            req.Body = io.NopCloser(bytes.NewReader(originalBody))
            req.ContentLength = int64(len(originalBody))
            if req.ContentLength >= 0 {
                req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
                req.TransferEncoding = nil
            }
        }
    } else {
        // 顶层不是对象（例如数组等），不处理
        req.Body = io.NopCloser(bytes.NewReader(originalBody))
        req.ContentLength = int64(len(originalBody))
        if req.ContentLength >= 0 {
            req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
            req.TransferEncoding = nil
        }
    }

    m.next.ServeHTTP(rw, req)
}

// 无额外辅助函数


