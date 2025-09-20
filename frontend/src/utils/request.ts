import Request, { RequestOptions } from 'luch-request'
import { ApiResponse, ApiError, ApiConfig } from './types'

// 默认配置
const defaultConfig: ApiConfig = {
  baseURL: 'http://localhost:8080/api',
  timeout: 10000,
  retryCount: 3,
  retryDelay: 1000
}

// 创建请求实例
const request = new Request()

// 获取基础URL
const getBaseURL = (): string => {
  // #ifdef H5
  return process.env.NODE_ENV === 'development'
    ? 'http://localhost:8080/api'
    : 'https://api.voicegenie.app/api'
  // #endif

  // #ifdef MP-WEIXIN
  return 'https://api.voicegenie.app/api'
  // #endif

  // #ifdef APP-PLUS
  return 'https://api.voicegenie.app/api'
  // #endif
}

// 请求配置
request.config = {
  baseURL: getBaseURL(),
  timeout: defaultConfig.timeout,
  dataType: 'json',
  responseType: 'text',

  // 自定义验证器，可以配置不符合请求格式的请求
  validateStatus: (statusCode: number) => {
    return statusCode >= 200 && statusCode < 300
  },

  // 自定义参数
  custom: {
    retry: true,
    retryCount: defaultConfig.retryCount,
    retryDelay: defaultConfig.retryDelay,
    loading: true,
    toast: true
  }
}

// 请求拦截器
request.interceptors.request.use(
  (config: RequestOptions) => {
    // 显示加载提示
    if (config.custom?.loading) {
      uni.showLoading({
        title: '加载中...',
        mask: true
      })
    }

    // 添加认证token
    const token = getToken()
    if (token) {
      config.header = {
        ...config.header,
        'Authorization': `Bearer ${token}`
      }
    }

    // 添加设备信息
    config.header = {
      ...config.header,
      'X-Device-Type': getDeviceType(),
      'X-App-Version': getAppVersion(),
      'X-Timestamp': Date.now().toString()
    }

    console.log('Request:', config)
    return config
  },
  (error: any) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: any) => {
    // 隐藏加载提示
    if (response.config.custom?.loading) {
      uni.hideLoading()
    }

    const { data, statusCode } = response

    // HTTP状态码检查
    if (statusCode !== 200) {
      const error = new ApiError(
        `请求失败 (${statusCode})`,
        statusCode,
        data
      )
      handleError(error, response.config)
      return Promise.reject(error)
    }

    // 业务状态码检查
    if (data.code !== 0) {
      const error = new ApiError(
        data.message || '请求失败',
        data.code,
        data.data
      )

      // 特殊错误处理
      if (data.code === 401) {
        handleUnauthorized()
      } else if (data.code === 403) {
        handleForbidden()
      } else {
        handleError(error, response.config)
      }

      return Promise.reject(error)
    }

    return data
  },
  (error: any) => {
    console.error('Response Error:', error)

    // 隐藏加载提示
    uni.hideLoading()

    // 网络错误处理
    if (!error.response) {
      const networkError = new ApiError('网络连接失败，请检查网络设置', -1)
      handleError(networkError, error.config)
      return Promise.reject(networkError)
    }

    // 服务器错误处理
    const serverError = new ApiError(
      '服务器错误，请稍后重试',
      error.response.statusCode || -1,
      error.response.data
    )

    handleError(serverError, error.config)
    return Promise.reject(serverError)
  }
)

// 错误处理
const handleError = (error: ApiError, config?: RequestOptions) => {
  console.error('API Error:', error)

  // 显示错误提示
  if (config?.custom?.toast !== false) {
    uni.showToast({
      title: error.message || '请求失败',
      icon: 'none',
      duration: 2000
    })
  }

  // 重试逻辑
  if (config?.custom?.retry && config.custom.retryCount! > 0) {
    config.custom.retryCount!--

    setTimeout(() => {
      request.request(config)
    }, config.custom.retryDelay || 1000)
  }
}

// 未授权处理
const handleUnauthorized = () => {
  // 清除本地token
  removeToken()

  // 跳转到登录页
  uni.reLaunch({
    url: '/pages/login/login'
  })

  uni.showToast({
    title: '登录已过期，请重新登录',
    icon: 'none'
  })
}

// 权限不足处理
const handleForbidden = () => {
  uni.showToast({
    title: '权限不足',
    icon: 'none'
  })
}

// Token管理
export const getToken = (): string | null => {
  try {
    return uni.getStorageSync('user_token')
  } catch (error) {
    console.error('获取token失败:', error)
    return null
  }
}

export const setToken = (token: string): void => {
  try {
    uni.setStorageSync('user_token', token)
  } catch (error) {
    console.error('保存token失败:', error)
  }
}

export const removeToken = (): void => {
  try {
    uni.removeStorageSync('user_token')
  } catch (error) {
    console.error('删除token失败:', error)
  }
}

// 获取设备类型
const getDeviceType = (): string => {
  // #ifdef H5
  return 'h5'
  // #endif

  // #ifdef MP-WEIXIN
  return 'mp-weixin'
  // #endif

  // #ifdef MP-ALIPAY
  return 'mp-alipay'
  // #endif

  // #ifdef APP-PLUS
  return 'app'
  // #endif

  return 'unknown'
}

// 获取应用版本
const getAppVersion = (): string => {
  try {
    // #ifdef APP-PLUS
    const systemInfo = uni.getSystemInfoSync()
    return systemInfo.appVersion || '1.0.0'
    // #endif

    return '1.0.0'
  } catch (error) {
    return '1.0.0'
  }
}

// 通用请求方法
export const http = {
  get<T = any>(url: string, params?: any, options?: RequestOptions): Promise<ApiResponse<T>> {
    return request.get(url, {
      params,
      ...options
    })
  },

  post<T = any>(url: string, data?: any, options?: RequestOptions): Promise<ApiResponse<T>> {
    return request.post(url, data, options)
  },

  put<T = any>(url: string, data?: any, options?: RequestOptions): Promise<ApiResponse<T>> {
    return request.put(url, data, options)
  },

  delete<T = any>(url: string, params?: any, options?: RequestOptions): Promise<ApiResponse<T>> {
    return request.delete(url, {
      params,
      ...options
    })
  },

  upload<T = any>(url: string, filePath: string, options?: {
    name?: string
    formData?: any
    onProgress?: (progress: number) => void
  }): Promise<ApiResponse<T>> {
    return new Promise((resolve, reject) => {
      const uploadTask = uni.uploadFile({
        url: getBaseURL() + url,
        filePath,
        name: options?.name || 'file',
        formData: options?.formData,
        header: {
          'Authorization': `Bearer ${getToken()}`,
          'X-Device-Type': getDeviceType(),
          'X-App-Version': getAppVersion(),
          'X-Timestamp': Date.now().toString()
        },
        success: (res) => {
          try {
            const data = JSON.parse(res.data)
            if (data.code === 0) {
              resolve(data)
            } else {
              reject(new ApiError(data.message || '上传失败', data.code, data.data))
            }
          } catch (error) {
            reject(new ApiError('响应解析失败', -1))
          }
        },
        fail: (error) => {
          reject(new ApiError('上传失败', -1, error))
        }
      })

      // 监听上传进度
      if (options?.onProgress) {
        uploadTask.onProgressUpdate((res) => {
          options.onProgress!(res.progress)
        })
      }
    })
  }
}

// 导出请求实例（用于特殊需求）
export default request