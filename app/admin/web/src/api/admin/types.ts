export interface PageQuery {
  page: number
  pageSize: number
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponseData {
  accessToken: string
}

export interface AdminUserDataItem {
  id: number
  username: string
  nickname: string
  password?: string
  email: string
  phone: string
  roles: string[]
  updatedAt: string
  createdAt: string
}

export type GetAdminUsersRequest = PageQuery &
  Partial<Pick<AdminUserDataItem, 'username' | 'nickname' | 'phone' | 'email'>> & {
    id?: string | number
  }
export type AdminUserCreateRequest = Pick<
  AdminUserDataItem,
  'username' | 'nickname' | 'email' | 'phone' | 'roles'
> & { password: string }
export type AdminUserUpdateRequest = Omit<AdminUserCreateRequest, 'password'> & {
  id: number
  password?: string
}

export interface MenuAuthDataItem {
  title: string
  authMark: string
}

export interface MenuDataItem {
  id?: number
  parentId?: number
  weight: number
  path: string
  title: string
  name?: string
  component?: string
  locale?: string
  icon?: string
  redirect?: string
  keepAlive?: boolean
  hideInMenu?: boolean
  isEnable?: boolean
  isMenu?: boolean
  isHide?: boolean
  isHideTab?: boolean
  link?: string
  isIframe?: boolean
  showBadge?: boolean
  showTextBadge?: string
  fixedTab?: boolean
  activePath?: string
  roles?: string[]
  isFullPage?: boolean
  authList?: MenuAuthDataItem[]
  target?: string
  url?: string
  updatedAt?: string
}

export interface RoleDataItem {
  id: number
  name: string
  sid: string
  updatedAt: string
  createdAt: string
}

export type GetRoleListRequest = PageQuery & Partial<Pick<RoleDataItem, 'sid' | 'name'>>
export type RoleCreateRequest = Pick<RoleDataItem, 'sid' | 'name'>
export type RoleUpdateRequest = RoleCreateRequest & { id: number }

export interface ApiDataItem {
  id: number
  name: string
  path: string
  method: string
  group: string
  menuIds: number[]
  updatedAt: string
  createdAt: string
}

export type GetApisRequest = PageQuery &
  Partial<Pick<ApiDataItem, 'group' | 'name' | 'path' | 'method'>>
export type ApiCreateRequest = Pick<ApiDataItem, 'group' | 'name' | 'path' | 'method' | 'menuIds'>
export type ApiUpdateRequest = ApiCreateRequest & { id: number }

export interface ListResponse<T> {
  list: T[]
  total: number
}

export interface GetApisResponseData extends ListResponse<ApiDataItem> {
  groups: string[]
}

export interface PermissionsData {
  list: string[]
}
