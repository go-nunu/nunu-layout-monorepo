import request from '@/utils/http'
import type { AdminUserCreateRequest, AdminUserUpdateRequest, ApiCreateRequest, ApiUpdateRequest, GetAdminUsersRequest, GetApisRequest, GetApisResponseData, GetRoleListRequest, ListResponse, MenuDataItem, PermissionsData, RoleCreateRequest, RoleDataItem, RoleUpdateRequest, AdminUserDataItem } from './types'

export function getRolesApi(params?: Partial<GetRoleListRequest>) {
  return request.get<ListResponse<RoleDataItem>>({ url: '/v1/admin/roles', params })
}
export function createRoleApi(params: RoleCreateRequest) {
  return request.post<any>({ url: '/v1/admin/role', params, showSuccessMessage: true })
}
export function updateRoleApi(params: RoleUpdateRequest) {
  return request.put<any>({ url: '/v1/admin/role', params, showSuccessMessage: true })
}
export function deleteRoleApi(params: any) {
  return request.del<any>({ url: '/v1/admin/role', params, showSuccessMessage: true })
}
export function getRolePermissionsApi(params: { role: string }) {
  return request.get<PermissionsData>({ url: '/v1/admin/role/permissions', params })
}
export function updateRolePermissionsApi(params: { role: string; list: string[] }) {
  return request.put<any>({ url: '/v1/admin/role/permissions', params, showSuccessMessage: true })
}

export function getAdminUsersApi(params?: Partial<GetAdminUsersRequest>) {
  return request.get<ListResponse<AdminUserDataItem>>({ url: '/v1/admin/users', params })
}
export function createAdminUserApi(params: AdminUserCreateRequest) {
  return request.post<any>({ url: '/v1/admin/user', params, showSuccessMessage: true })
}
export function updateAdminUserApi(params: AdminUserUpdateRequest) {
  return request.put<any>({ url: '/v1/admin/user', params, showSuccessMessage: true })
}
export function deleteAdminUserApi(params: any) {
  return request.del<any>({ url: '/v1/admin/user', params, showSuccessMessage: true })
}

export function getMenusApi() {
  return request.get<any>({ url: '/v1/menus' })
}
export function getAdminMenusApi(params?: any) {
  return request.get<ListResponse<MenuDataItem>>({ url: '/v1/admin/menus', params })
}
export function createMenuApi(params: Omit<MenuDataItem, 'id'>) {
  return request.post<any>({ url: '/v1/admin/menu', params, showSuccessMessage: true })
}
export function updateMenuApi(params: MenuDataItem & { id: number }) {
  return request.put<any>({ url: '/v1/admin/menu', params, showSuccessMessage: true })
}
export function deleteMenusApi(params: any) {
  return request.del<any>({ url: '/v1/admin/menu', params, showSuccessMessage: true })
}

export function getAdminApiApi(params?: Partial<GetApisRequest>) {
  return request.get<GetApisResponseData>({ url: '/v1/admin/apis', params })
}
export function createAdminApiApi(params: ApiCreateRequest) {
  return request.post<any>({ url: '/v1/admin/api', params, showSuccessMessage: true })
}
export function updateAdminApiApi(params: ApiUpdateRequest) {
  return request.put<any>({ url: '/v1/admin/api', params, showSuccessMessage: true })
}
export function deleteAdminApiApi(params: any) {
  return request.del<any>({ url: '/v1/admin/api', params, showSuccessMessage: true })
}
