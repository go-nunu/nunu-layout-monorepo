<template>
  <div class="admin-page art-full-height">
    <ElCard shadow="never" class="search-card">
      <ElForm :model="searchForm" label-width="82px">
        <ElRow :gutter="16">
          <ElCol :xs="24" :md="8">
            <ElFormItem label="角色标识">
              <ElInput v-model="searchForm.sid" clearable />
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8">
            <ElFormItem label="角色名称">
              <ElInput v-model="searchForm.name" clearable />
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8" class="search-actions">
            <ElButton type="primary" :loading="loading" @click="loadData">查询</ElButton>
            <ElButton :loading="loading" @click="resetSearch">重置</ElButton>
          </ElCol>
        </ElRow>
      </ElForm>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>角色列表</span>
          <ElSpace>
            <ElButton type="primary" @click="openCreate">新增角色</ElButton>
            <ElButton :loading="loading" @click="loadData">刷新</ElButton>
          </ElSpace>
        </div>
      </template>

      <ElTable v-loading="loading" :data="rows" height="100%" border stripe>
        <ElTableColumn prop="id" label="ID" width="90" />
        <ElTableColumn prop="sid" label="角色标识" min-width="160" />
        <ElTableColumn prop="name" label="角色名称" min-width="180" />
        <ElTableColumn label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <ElButton link type="primary" @click="openPermission(row)">权限</ElButton>
            <ElButton link type="primary" @click="openEdit(row)">编辑</ElButton>
            <ElButton link type="danger" @click="remove(row)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <ElPagination
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.pageSize"
        class="pagination"
        background
        layout="total, sizes, prev, pager, next"
        :total="pagination.total"
        @change="loadData"
      />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="form.id ? '编辑角色' : '新增角色'"
      width="480px"
      align-center
      destroy-on-close
      @closed="clearFormValidate"
    >
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
        <ElFormItem label="角色标识" prop="sid">
          <ElInput
            v-model="form.sid"
            :disabled="!!form.id"
            placeholder="唯一标识，创建后不可修改"
          />
        </ElFormItem>
        <ElFormItem label="角色名称" prop="name">
          <ElInput v-model="form.name" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="submit">确定</ElButton>
      </template>
    </ElDialog>

    <ElDialog
      v-model="permissionVisible"
      :title="`权限配置：${permissionRole.name || ''}`"
      width="760px"
      align-center
      destroy-on-close
    >
      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="接口权限" name="api">
          <ElScrollbar max-height="56vh">
            <ElTree
              ref="apiTreeRef"
              :data="apiTree"
              show-checkbox
              node-key="key"
              default-expand-all
              :props="treeProps"
            >
              <template #default="{ data }">
                <ElTag v-if="data.group && data.group !== data.title" size="small" class="mr-2">
                  {{ data.group }}
                </ElTag>
                <span>{{ data.title }}</span>
              </template>
            </ElTree>
          </ElScrollbar>
        </ElTabPane>
        <ElTabPane label="菜单权限" name="menu">
          <ElScrollbar max-height="56vh">
            <ElTree
              ref="menuTreeRef"
              :data="menuTree"
              show-checkbox
              node-key="key"
              default-expand-all
              :props="treeProps"
            />
          </ElScrollbar>
        </ElTabPane>
      </ElTabs>

      <template #footer>
        <ElButton @click="permissionVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="submitPermission">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElMessage,
    ElMessageBox,
    type ElTree,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import {
    createRoleApi,
    deleteRoleApi,
    getAdminApiApi,
    getAdminMenusApi,
    getRolePermissionsApi,
    getRolesApi,
    updateRoleApi,
    updateRolePermissionsApi
  } from '@/api/admin/permission'
  import type {
    ApiDataItem,
    GetRoleListRequest,
    MenuDataItem,
    RoleCreateRequest,
    RoleDataItem,
    RoleUpdateRequest
  } from '@/api/admin/types'
  import { formatMenuTitle } from '@/utils/router'

  defineOptions({ name: 'AdminRole' })

  interface RoleForm {
    id?: number
    sid: string
    name: string
  }

  interface TreeNode {
    key: string
    title: string
    group?: string
    children?: TreeNode[]
  }

  type ListPayload<T> =
    | T[]
    | {
        list?: T[]
        records?: T[]
        items?: T[]
        total?: number
        count?: number
      }

  const loading = ref(false)
  const submitting = ref(false)
  const rows = ref<RoleDataItem[]>([])
  const dialogVisible = ref(false)
  const permissionVisible = ref(false)
  const activeTab = ref('api')
  const formRef = ref<FormInstance>()
  const apiTreeRef = ref<InstanceType<typeof ElTree>>()
  const menuTreeRef = ref<InstanceType<typeof ElTree>>()
  const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
  const searchForm = reactive<Partial<GetRoleListRequest>>({ sid: '', name: '' })
  const form = reactive<RoleForm>(createDefaultForm())
  const permissionRole = ref<Partial<RoleDataItem>>({})
  const apiTree = ref<TreeNode[]>([])
  const menuTree = ref<TreeNode[]>([])
  const treeProps = { label: 'title', children: 'children' } as const
  const rules: FormRules<RoleForm> = {
    sid: [{ required: true, message: '请输入角色标识', trigger: 'blur' }],
    name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }]
  }

  function createDefaultForm(): RoleForm {
    return {
      id: undefined,
      sid: '',
      name: ''
    }
  }

  function normalizeList<T>(data: ListPayload<T>): T[] {
    if (Array.isArray(data)) {
      return data
    }

    return data.list || data.records || data.items || []
  }

  function normalizeTotal<T>(data: ListPayload<T>, list: T[]) {
    if (Array.isArray(data)) {
      return list.length
    }

    return Number(data.total ?? data.count ?? list.length)
  }

  function resetForm(data: Partial<RoleForm> = {}) {
    Object.assign(form, createDefaultForm(), data)
    nextTick(clearFormValidate)
  }

  function clearFormValidate() {
    formRef.value?.clearValidate()
  }

  async function loadData() {
    loading.value = true

    try {
      const data = await getRolesApi({
        ...searchForm,
        page: pagination.current,
        pageSize: pagination.pageSize
      })
      const list = normalizeList(data)
      rows.value = list
      pagination.total = normalizeTotal(data, list)
    } finally {
      loading.value = false
    }
  }

  async function resetSearch() {
    searchForm.sid = ''
    searchForm.name = ''
    pagination.current = 1
    await loadData()
  }

  function openCreate() {
    resetForm()
    dialogVisible.value = true
  }

  function openEdit(row: RoleDataItem) {
    resetForm(row)
    dialogVisible.value = true
  }

  async function remove(row: RoleDataItem) {
    await ElMessageBox.confirm(`确定删除角色「${row.name || row.sid}」吗？`, '删除确认', {
      type: 'warning'
    })
    await deleteRoleApi({ id: row.id, sid: row.sid })
    ElMessage.success('删除成功')
    await loadData()
  }

  async function submit() {
    if (!formRef.value) {
      throw new Error('角色表单未初始化')
    }

    await formRef.value.validate()
    submitting.value = true

    try {
      if (form.id) {
        const payload: RoleUpdateRequest = { id: form.id, sid: form.sid, name: form.name }
        await updateRoleApi(payload)
      } else {
        const payload: RoleCreateRequest = { sid: form.sid, name: form.name }
        await createRoleApi(payload)
      }

      ElMessage.success('提交成功')
      dialogVisible.value = false
      await loadData()
    } finally {
      submitting.value = false
    }
  }

  function menuToTree(list: MenuDataItem[]): TreeNode[] {
    const map = new Map<number, TreeNode & MenuDataItem>()
    const result: TreeNode[] = []

    list.forEach((item) => {
      if (item.id === undefined) return

      map.set(item.id, {
        ...item,
        key: '',
        title: formatMenuTitle(item.title || item.name || item.path),
        children: []
      })
    })

    map.forEach((node) => {
      node.key = `menu:${resolveFullMenuPath(node, map)},read`
    })

    map.forEach((node) => {
      const parent = node.parentId ? map.get(node.parentId) : undefined
      if (parent) {
        parent.children?.push(node)
      } else {
        result.push(node)
      }
    })

    return result
  }

  function resolveFullMenuPath(
    item: Pick<MenuDataItem, 'path' | 'parentId'>,
    map: Map<number, MenuDataItem>
  ): string {
    const currentPath = item.path || ''
    if (
      currentPath.startsWith('http://') ||
      currentPath.startsWith('https://') ||
      currentPath.startsWith('/')
    ) {
      return currentPath
    }

    if (!item.parentId) {
      return `/${currentPath.replace(/^\//, '')}`
    }

    const parent = map.get(item.parentId)
    if (!parent) {
      return `/${currentPath.replace(/^\//, '')}`
    }

    const parentPath = resolveFullMenuPath(parent, map).replace(/\/$/, '')
    const childPath = currentPath.replace(/^\//, '')
    return parentPath ? `${parentPath}/${childPath}` : `/${childPath}`
  }

  function apiToTree(list: ApiDataItem[]): TreeNode[] {
    const groups = new Map<string, TreeNode>()

    list.forEach((item) => {
      const group = item.group || '默认分组'
      if (!groups.has(group)) {
        groups.set(group, { key: `group:${group}`, title: group, children: [] })
      }

      groups.get(group)?.children?.push({
        key: `api:${item.path},${item.method}`,
        title: `${item.name || item.path} ${item.method || ''}`,
        group
      })
    })

    return [...groups.values()]
  }

  async function openPermission(row: RoleDataItem) {
    permissionRole.value = row
    activeTab.value = 'api'
    const [menus, rolePerms, apis] = await Promise.all([
      getAdminMenusApi({}),
      getRolePermissionsApi({ role: row.sid }),
      getAdminApiApi({ page: 1, pageSize: 999 })
    ])

    menuTree.value = menuToTree(normalizeList(menus))
    apiTree.value = apiToTree(normalizeList(apis))
    permissionVisible.value = true

    await nextTick()
    const checked = normalizeList(rolePerms)
    apiTreeRef.value?.setCheckedKeys(checked.filter((item) => item.startsWith('api:')))
    menuTreeRef.value?.setCheckedKeys(checked.filter((item) => item.startsWith('menu:')))
  }

  async function submitPermission() {
    submitting.value = true

    try {
      const list = [
        ...(apiTreeRef.value?.getCheckedKeys(false) || []),
        ...(menuTreeRef.value?.getCheckedKeys(false) || [])
      ]
        .map(String)
        .filter((key) => key.includes(','))

      await updateRolePermissionsApi({ role: permissionRole.value.sid || '', list })
      ElMessage.success('权限已更新')
      permissionVisible.value = false
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

  .search-card {
    flex: none;
  }

  .art-table-card {
    min-height: 0;
  }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .search-actions {
    display: flex;
    align-items: flex-start;
  }

  .pagination {
    justify-content: flex-end;
    margin-top: 16px;
  }

  .permission-tabs {
    min-height: 360px;
  }

  .mr-2 {
    margin-right: 8px;
  }
</style>
