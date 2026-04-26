import request from '@/utils/http'

export interface LoginParams {
  username?: string
  userName?: string
  password?: string
  mobile?: string
  code?: string
  type?: string
}

export function loginApi(params: LoginParams) {
  return request.post<any>({
    url: '/v1/login',
    params,
    showErrorMessage: true
  })
}

export function getAdminUserInfoApi() {
  return request.get<any>({ url: '/v1/admin/user' })
}
