import { AppRouteRecord } from '@/types/router'

export const adminRoutes: AppRouteRecord = {
  path: '/admin',
  name: 'AdminManage',
  component: '/index/index',
  meta: {
    title: '权限管理',
    icon: 'ri:shield-user-line',
    roles: ['admin']
  },
  children: [
    {
      path: 'user',
      name: 'AdminUser',
      component: '/admin/user',
      meta: { title: '用户管理', icon: 'ri:user-line', keepAlive: true }
    },
    {
      path: 'role',
      name: 'AdminRole',
      component: '/admin/role',
      meta: { title: '角色管理', icon: 'ri:user-settings-line', keepAlive: true }
    },
    {
      path: 'menu',
      name: 'AdminMenu',
      component: '/admin/menu',
      meta: { title: '菜单管理', icon: 'ri:menu-line', keepAlive: true }
    },
    {
      path: 'api',
      name: 'AdminApi',
      component: '/admin/api',
      meta: { title: '接口管理', icon: 'ri:terminal-window-line', keepAlive: true }
    }
  ]
}
