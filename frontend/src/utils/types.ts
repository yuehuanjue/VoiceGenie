import { RequestOptions } from 'luch-request'

// API响应接口
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp: number
}

// 分页响应接口
export interface PaginatedResponse<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
  hasMore: boolean
}

// 用户信息接口
export interface UserInfo {
  id: string
  nickname: string
  avatar: string
  phone?: string
  email?: string
  loginType: 'phone' | 'wechat' | 'guest'
  createdAt: number
  updatedAt: number
}

// 消息接口
export interface Message {
  id: string
  conversationId: string
  type: 'user' | 'ai' | 'system'
  content: string
  audioUrl?: string
  audioDuration?: number
  timestamp: number
  status?: 'sending' | 'sent' | 'failed'
}

// 对话接口
export interface Conversation {
  id: string
  title: string
  lastMessage: string
  messageCount: number
  duration: number
  createdAt: number
  updatedAt: number
}

// 语音识别结果
export interface ASRResult {
  text: string
  confidence: number
  language: string
  duration: number
}

// TTS请求参数
export interface TTSRequest {
  text: string
  voice?: string
  speed?: number
  pitch?: number
  volume?: number
}

// TTS响应
export interface TTSResult {
  audioUrl: string
  duration: number
  text: string
}

// 聊天请求参数
export interface ChatRequest {
  message: string
  conversationId?: string
  context?: any
}

// 聊天响应
export interface ChatResponse {
  reply: string
  conversationId: string
  audioUrl?: string
  suggestions?: string[]
}

// API错误类
export class ApiError extends Error {
  public code: number
  public data?: any

  constructor(message: string, code: number = -1, data?: any) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.data = data
  }
}

// 请求配置
export interface ApiConfig {
  baseURL: string
  timeout: number
  retryCount: number
  retryDelay: number
}

// 上传进度回调
export type UploadProgressCallback = (progress: number, total: number) => void

// 文件上传响应
export interface UploadResponse {
  url: string
  filename: string
  size: number
  type: string
}