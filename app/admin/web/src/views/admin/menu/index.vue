<template>
  <div class="admin-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>菜单管理</span>
          <ElSpace>
            <ElButton type="primary" @click="openCreate()">新增菜单</ElButton>
            <ElButton :loading="loading" @click="loadData">刷新</ElButton>
          </ElSpace>
        </div>
      </template>

      <ElTable
        v-loading="loading"
        :data="rows"
        row-key="id"
        height="100%"
        border
        stripe
        default-expand-all
      >
        <ElTableColumn label="菜单名称" min-width="180">
          <template #default="{ row }">{{ formatMenuTitle(row.title) }}</template>
        </ElTableColumn>
        <ElTableColumn prop="path" label="路由路径" min-width="180" />
        <ElTableColumn prop="name" label="权限标识" min-width="150" />
        <ElTableColumn prop="component" label="组件" min-width="180" />
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="row.isEnable === false ? 'info' : 'success'">
              {{ row.isEnable === false ? '禁用' : '启用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="隐藏" width="90">
          <template #default="{ row }">
            <ElTag :type="isMenuHidden(row) ? 'warning' : 'success'">
              {{ isMenuHidden(row) ? '是' : '否' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="210" fixed="right">
          <template #default="{ row }">
            <ElButton link type="primary" @click="openCreate(row)">新增子级</ElButton>
            <ElButton link type="primary" @click="openEdit(row)">编辑</ElButton>
            <ElButton link type="danger" @click="remove(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="860px"
      align-center
      destroy-on-close
      @closed="clearFormValidate"
    >
      <ElScrollbar max-height="68vh">
        <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px" class="menu-form">
          <ElFormItem label="父级菜单" class="menu-form__half">
            <ElTreeSelect
              v-model="form.parentId"
              :data="treeOptions"
              node-key="id"
              :props="treeProps"
              check-strictly
              clearable
              class="w-full"
            />
          </ElFormItem>

          <ElDivider content-position="left">基础配置</ElDivider>
          <div class="menu-form__grid">
            <ElFormItem label="菜单名称" prop="title">
              <ElInput v-model="form.title" placeholder="请输入菜单名称" />
            </ElFormItem>
            <ElFormItem label="路由地址" prop="path">
              <ElInput v-model="form.path" placeholder="如：/dashboard 或 console" />
            </ElFormItem>
            <ElFormItem label="权限标识" prop="name">
              <ElInput v-model="form.name" placeholder="如：User" />
            </ElFormItem>
            <ElFormItem label="组件路径">
              <ElInput v-model="form.component" placeholder="如：/system/user 或留空" />
            </ElFormItem>
            <ElFormItem label="图标">
              <ElInput v-model="form.icon" placeholder="如：ri:user-line" />
            </ElFormItem>
            <ElFormItem label="菜单排序">
              <ElInputNumber
                v-model="form.weight"
                :min="0"
                controls-position="right"
                class="w-full"
              />
            </ElFormItem>
            <ElFormItem label="本地化">
              <ElInput v-model="form.locale" placeholder="如：menu.admin.user" />
            </ElFormItem>
            <ElFormItem label="重定向">
              <ElInput v-model="form.redirect" placeholder="如：/dashboard/workplace" />
            </ElFormItem>
          </div>

          <ElDivider content-position="left">访问与展示</ElDivider>
          <div class="menu-form__grid">
            <ElFormItem label="角色权限">
              <ElSelect
                v-model="form.roles"
                multiple
                filterable
                allow-create
                default-first-option
                class="w-full"
                placeholder="输入角色标识后回车，如：R_SUPER"
              >
                <ElOption
                  v-for="role in roleOptions"
                  :key="role.sid"
                  :label="`${role.name}（${role.sid}）`"
                  :value="role.sid"
                />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="外部链接">
              <ElInput v-model="form.link" placeholder="如：https://www.example.com" />
            </ElFormItem>
            <ElFormItem label="iframe URL">
              <ElInput v-model="form.url" placeholder="旧版 iframe 地址，可留空" />
            </ElFormItem>
            <ElFormItem label="打开方式">
              <ElSelect
                v-model="form.target"
                clearable
                class="w-full"
                placeholder="请选择外链打开方式"
              >
                <ElOption label="_blank" value="_blank" />
                <ElOption label="_self" value="_self" />
                <ElOption label="_parent" value="_parent" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="文本徽章">
              <ElInput v-model="form.showTextBadge" placeholder="如：New、Hot" />
            </ElFormItem>
            <ElFormItem label="激活路径">
              <ElInput v-model="form.activePath" placeholder="如：/system/user" />
            </ElFormItem>
          </div>

          <div class="menu-form__switches">
            <ElFormItem label="是否启用"><ElSwitch v-model="form.isEnable" /></ElFormItem>
            <ElFormItem label="页面缓存"><ElSwitch v-model="form.keepAlive" /></ElFormItem>
            <ElFormItem label="隐藏菜单"><ElSwitch v-model="form.isHide" /></ElFormItem>
            <ElFormItem label="是否内嵌"><ElSwitch v-model="form.isIframe" /></ElFormItem>
            <ElFormItem label="显示徽章"><ElSwitch v-model="form.showBadge" /></ElFormItem>
            <ElFormItem label="固定标签"><ElSwitch v-model="form.fixedTab" /></ElFormItem>
            <ElFormItem label="标签隐藏"><ElSwitch v-model="form.isHideTab" /></ElFormItem>
            <ElFormItem label="全屏页面"><ElSwitch v-model="form.isFullPage" /></ElFormItem>
          </div>

          <ElDivider content-position="left">按钮权限</ElDivider>
          <div class="auth-list">
            <div v-for="(_, index) in form.authList" :key="index" class="auth-list__row">
              <ElInput v-model="form.authList[index].title" placeholder="权限名称，如：新增" />
              <ElInput v-model="form.authList[index].authMark" placeholder="权限标识，如：add" />
              <ElButton type="danger" text @click="removeAuth(index)">删除</ElButton>
            </div>
            <ElButton type="primary" text @click="addAuth">添加按钮权限</ElButton>
          </div>
        </ElForm>
      </ElScrollbar>

      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="submit">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
  import {
    createMenuApi,
    deleteMenusApi,
    getAdminMenusApi,
    getRolesApi,
    updateMenuApi
  } from '@/api/admin/permission'
  import { formatMenuTitle } from '@/utils/router'
  import type { MenuAuthDataItem, MenuDataItem, RoleDataItem } from '@/api/admin/types'

  defineOptions({ name: 'AdminMenu' })

  type MenuTreeItem = MenuDataItem & {
    title: string
    displayTitle: string
    children: MenuTreeItem[]
  }

  type ListPayload<T> =
    | T[]
    | {
        list?: T[]
        records?: T[]
        items?: T[]
      }

  type MenuForm = MenuDataItem & {
    roles: string[]
    authList: MenuAuthDataItem[]
  }

  const loading = ref(false)
  const submitting = ref(false)
  const rows = ref<MenuTreeItem[]>([])
  const roleOptions = ref<RoleDataItem[]>([])
  const dialogVisible = ref(false)
  const formRef = ref<FormInstance>()
  const form = reactive<MenuForm>(createDefaultForm())

  const treeProps = { label: 'displayTitle', children: 'children' } as const
  const rules: FormRules<MenuForm> = {
    title: [{ required: true, message: '请输入菜单名称', trigger: 'blur' }],
    path: [{ required: true, message: '请输入路由地址', trigger: 'blur' }],
    name: [{ required: true, message: '请输入权限标识', trigger: 'blur' }]
  }

  const treeOptions = computed(() =>
    form.id ? excludeMenuBranch(rows.value, form.id) : rows.value
  )
  const dialogTitle = computed(() => (form.id ? '编辑菜单' : '新增菜单'))

  function createDefaultForm(): MenuForm {
    return {
      id: undefined,
      parentId: undefined,
      title: '',
      path: '',
      name: '',
      component: '',
      icon: '',
      weight: 0,
      locale: '',
      redirect: '',
      url: '',
      link: '',
      target: '',
      activePath: '',
      showTextBadge: '',
      isEnable: true,
      isMenu: true,
      keepAlive: false,
      hideInMenu: false,
      isHide: false,
      isHideTab: false,
      isIframe: false,
      showBadge: false,
      fixedTab: false,
      isFullPage: false,
      roles: [],
      authList: []
    }
  }

  function normalizeList<T>(data: ListPayload<T>): T[] {
    if (Array.isArray(data)) {
      return data
    }

    return data.list || data.records || data.items || []
  }

  function toTree(list: MenuDataItem[]): MenuTreeItem[] {
    const map = new Map<number, MenuTreeItem>()
    const result: MenuTreeItem[] = []

    list.forEach((item) => {
      if (item.id === undefined) return

      map.set(item.id, {
        ...item,
        title: item.title || item.name || item.path,
        displayTitle: formatMenuTitle(item.title || item.name || item.path),
        children: []
      })
    })

    map.forEach((node) => {
      const parent = node.parentId ? map.get(node.parentId) : undefined

      if (parent) {
        parent.children.push(node)
      } else {
        result.push(node)
      }
    })

    return result
  }

  function excludeMenuBranch(nodes: MenuTreeItem[], excludedID: number): MenuTreeItem[] {
    return nodes
      .filter((node) => node.id !== excludedID)
      .map((node) => ({
        ...node,
        children: excludeMenuBranch(node.children, excludedID)
      }))
  }

  function isMenuHidden(row: Pick<MenuDataItem, 'hideInMenu' | 'isHide'>): boolean {
    return Boolean(row.isHide || row.hideInMenu)
  }

  function pickFormData(data: Partial<MenuDataItem>): Partial<MenuForm> {
    return {
      id: data.id,
      parentId: data.parentId,
      title: data.title,
      path: data.path,
      name: data.name,
      component: data.component,
      icon: data.icon,
      weight: data.weight,
      locale: data.locale,
      redirect: data.redirect,
      url: data.url,
      link: data.link,
      target: data.target,
      activePath: data.activePath,
      showTextBadge: data.showTextBadge,
      isEnable: data.isEnable ?? true,
      isMenu: data.isMenu ?? true,
      keepAlive: data.keepAlive,
      hideInMenu: isMenuHidden(data),
      isHide: isMenuHidden(data),
      isHideTab: data.isHideTab,
      isIframe: data.isIframe,
      showBadge: data.showBadge,
      fixedTab: data.fixedTab,
      isFullPage: data.isFullPage,
      roles: [...(data.roles || [])],
      authList: (data.authList || []).map((item) => ({ ...item }))
    }
  }

  function resetForm(data: Partial<MenuForm> = {}) {
    Object.assign(form, createDefaultForm(), data)
    nextTick(clearFormValidate)
  }

  function clearFormValidate() {
    formRef.value?.clearValidate()
  }

  function addAuth() {
    form.authList.push({ title: '', authMark: '' })
  }

  function removeAuth(index: number) {
    form.authList.splice(index, 1)
  }

  function createPayload(): MenuDataItem {
    const isHide = Boolean(form.isHide)
    return {
      ...form,
      hideInMenu: isHide,
      isHide,
      roles: [...form.roles],
      authList: form.authList
        .filter((item) => item.title || item.authMark)
        .map((item) => ({ title: item.title, authMark: item.authMark }))
    }
  }

  async function loadRoleOptions() {
    const data = await getRolesApi({ page: 1, pageSize: 999 })
    roleOptions.value = normalizeList(data)
  }

  async function loadData() {
    loading.value = true

    try {
      const [menus] = await Promise.all([getAdminMenusApi({}), loadRoleOptions()])
      rows.value = toTree(normalizeList(menus))
    } finally {
      loading.value = false
    }
  }

  function openCreate(parent?: MenuTreeItem) {
    resetForm(parent ? { parentId: parent.id } : {})
    dialogVisible.value = true
  }

  function openEdit(row: MenuTreeItem) {
    resetForm(pickFormData(row))
    dialogVisible.value = true
  }

  async function remove(row: MenuTreeItem) {
    if (row.children.length) {
      ElMessage.warning('该菜单仍有子菜单，请先删除或移动子菜单')
      return
    }
    await ElMessageBox.confirm(
      `确定删除菜单「${formatMenuTitle(row.title || row.name || '')}」吗？`,
      '删除确认',
      {
        type: 'warning'
      }
    )
    await deleteMenusApi({ id: row.id })
    ElMessage.success('删除成功')
    await loadData()
  }

  async function submit() {
    if (!formRef.value) {
      throw new Error('菜单表单未初始化')
    }

    await formRef.value.validate()
    submitting.value = true

    try {
      const payload = createPayload()
      if (payload.id) {
        await updateMenuApi({ ...payload, id: payload.id })
      } else {
        await createMenuApi(payload)
      }

      ElMessage.success('提交成功')
      dialogVisible.value = false
      await loadData()
    } finally {
      submitting.value = false
    }
  }

  onMounted(loadData)
</script>

<style scoped lang="scss">
  .admin-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .art-table-card {
    min-height: 0;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .menu-form {
    padding: 8px 8px 0 0;

    &__grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      column-gap: 24px;
    }

    &__half {
      max-width: calc(50% - 12px);
    }

    &__switches {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      column-gap: 16px;
    }
  }

  .auth-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding-left: 100px;

    &__row {
      display: grid;
      grid-template-columns: 1fr 1fr auto;
      gap: 10px;
      align-items: center;
    }
  }

  .w-full {
    width: 100%;
  }

  @media (width <= 768px) {
    .menu-form {
      &__grid,
      &__switches {
        grid-template-columns: 1fr;
      }

      &__half {
        max-width: none;
      }
    }

    .auth-list {
      padding-left: 0;

      &__row {
        grid-template-columns: 1fr;
      }
    }
  }
</style>
