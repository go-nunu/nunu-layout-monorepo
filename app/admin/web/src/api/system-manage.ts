import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'
import type { MenuDataItem } from './admin/types'

// 获取用户列表
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/user/list',
    params
  })
}

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: '/api/role/list',
    params
  })
}

// 获取菜单列表
export function fetchGetMenuList() {
  return request
    .get<{ list: MenuDataItem[] }>({
      url: '/v1/menus'
    })
    .then((data) => buildRouteTree(data.list || []))
}

function buildRouteTree(list: MenuDataItem[]): AppRouteRecord[] {
  const map = new Map<number, AppRouteRecord>()
  const result: AppRouteRecord[] = []

  list.forEach((item) => {
    if (item.id === undefined) return
    map.set(item.id, menuToRouteRecord(item))
  })

  list.forEach((item) => {
    if (item.id === undefined) return
    const route = map.get(item.id)
    if (!route) return

    const parent = item.parentId ? map.get(item.parentId) : undefined
    if (parent) {
      parent.children = parent.children || []
      parent.children.push(route)
    } else {
      result.push(route)
    }
  })

  map.forEach((route) => {
    if (route.children?.length === 0) {
      delete route.children
    }
  })

  return result
}

function menuToRouteRecord(item: MenuDataItem): AppRouteRecord {
  return {
    id: item.id,
    path: item.path,
    name: item.name,
    component: item.component,
    redirect: item.redirect,
    meta: {
      title: item.title || item.name || item.path,
      icon: item.icon,
      showBadge: item.showBadge,
      showTextBadge: item.showTextBadge,
      isHide: item.isHide || item.hideInMenu,
      isHideTab: item.isHideTab,
      link: item.link,
      isIframe: item.isIframe,
      keepAlive: item.keepAlive,
      authList: item.authList,
      roles: item.roles,
      fixedTab: item.fixedTab,
      activePath: item.activePath,
      isFullPage: item.isFullPage
    },
    children: []
  }
}
