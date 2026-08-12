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
      width="1040px"
      align-center
      destroy-on-close
      @closed="closePermission"
    >
      <div v-loading="permissionLoading" class="permission-editor">
        <ElAlert type="info" :closable="false" show-icon>
          <template #title>
            勾选菜单会自动勾选并锁定它及子菜单关联的 API；取消菜单后，关联 API 会自动释放。
          </template>
        </ElAlert>

        <div class="permission-summary" aria-live="polite">
          <span>已选 {{ selectedMenuCount }} 项菜单</span>
          <span class="permission-summary__dot"></span>
          <span>已选 {{ selectedApiCount }} 个 API</span>
          <span v-if="lastAutoLinkedCount" class="permission-summary__linked">
            本次自动关联 {{ lastAutoLinkedCount }} 个
          </span>
        </div>

        <div class="permission-grid">
          <section class="permission-panel" aria-labelledby="menu-permission-title">
            <header class="permission-panel__header">
              <div>
                <h3 id="menu-permission-title">菜单权限</h3>
                <p>按业务层级选择可见页面</p>
              </div>
              <ElTag effect="plain">{{ selectedMenuCount }}</ElTag>
            </header>
            <ElScrollbar max-height="48vh" class="permission-panel__body">
              <ElTree
                ref="menuTreeRef"
                :data="menuTree"
                show-checkbox
                node-key="key"
                default-expand-all
                :props="treeProps"
                empty-text="暂无菜单数据"
                @check="handleMenuCheck"
              >
                <template #default="{ data }">
                  <div class="tree-node">
                    <span class="tree-node__title">{{ data.title }}</span>
                    <ElTag v-if="data.linkedApiCount" size="small" type="success" effect="plain">
                      {{ data.linkedApiCount }} API
                    </ElTag>
                  </div>
                </template>
              </ElTree>
            </ElScrollbar>
          </section>

          <section class="permission-panel" aria-labelledby="api-permission-title">
            <header class="permission-panel__header">
              <div>
                <h3 id="api-permission-title">API 权限</h3>
                <p>按多级分类管理后端访问范围</p>
              </div>
              <ElTag effect="plain">{{ selectedApiCount }}</ElTag>
            </header>
            <ElScrollbar max-height="56vh">
              <ElTree
                ref="apiTreeRef"
                :data="apiTree"
                show-checkbox
                node-key="key"
                default-expand-all
                :props="treeProps"
                empty-text="暂无 API 数据"
                @check="handleApiCheck"
              >
                <template #default="{ data }">
                  <div class="tree-node" :class="{ 'tree-node--api': data.kind === 'api' }">
                    <span class="tree-node__title">{{ data.title }}</span>
                    <ElTag
                      v-if="data.method"
                      size="small"
                      :type="methodTagType(data.method)"
                      effect="plain"
                    >
                      {{ data.method }}
                    </ElTag>
                    <ElTooltip
                      v-if="data.linkedMenuNames?.length"
                      :content="`关联菜单：${data.linkedMenuNames.join('、')}`"
                      placement="top"
                    >
                      <ElIcon class="tree-node__link" aria-label="已关联菜单">
                        <Link />
                      </ElIcon>
                    </ElTooltip>
                  </div>
                </template>
              </ElTree>
            </ElScrollbar>
          </section>
        </div>
      </div>

      <template #footer>
        <ElButton @click="permissionVisible = false">取消</ElButton>
        <ElButton
          type="primary"
          :loading="submitting"
          :disabled="permissionLoading"
          @click="submitPermission"
        >
          保存权限
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { Link } from '@element-plus/icons-vue'
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
    kind: 'menu' | 'category' | 'api'
    menuId?: number
    method?: string
    linkedApiCount?: number
    linkedMenuNames?: string[]
    disabled?: boolean
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
  const permissionLoading = ref(false)
  const rows = ref<RoleDataItem[]>([])
  const dialogVisible = ref(false)
  const permissionVisible = ref(false)
  const formRef = ref<FormInstance>()
  const apiTreeRef = ref<InstanceType<typeof ElTree>>()
  const menuTreeRef = ref<InstanceType<typeof ElTree>>()
  const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
  const searchForm = reactive<Partial<GetRoleListRequest>>({ sid: '', name: '' })
  const form = reactive<RoleForm>(createDefaultForm())
  const permissionRole = ref<Partial<RoleDataItem>>({})
  const apiTree = ref<TreeNode[]>([])
  const menuTree = ref<TreeNode[]>([])
  const selectedMenuCount = ref(0)
  const selectedApiCount = ref(0)
  const lastAutoLinkedCount = ref(0)
  const menuApiKeys = new Map<number, Set<string>>()
  const autoCheckedApiKeys = new Set<string>()
  const lockedApiKeys = new Set<string>()
  let previousCheckedMenuIDs = new Set<number>()
  let permissionRequestId = 0
  const treeProps = { label: 'title', children: 'children', disabled: 'disabled' } as const
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
        kind: 'menu',
        menuId: item.id,
        linkedApiCount: menuApiKeys.get(item.id)?.size || 0,
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

    const collectLinkedApiKeys = (node: TreeNode): Set<string> => {
      const keys = new Set(menuApiKeys.get(node.menuId || 0) || [])
      node.children?.forEach((child) => {
        collectLinkedApiKeys(child).forEach((key) => keys.add(key))
      })
      node.linkedApiCount = keys.size
      return keys
    }
    result.forEach(collectLinkedApiKeys)

    return result
  }

  function resolveFullMenuPath(
    item: Pick<MenuDataItem, 'id' | 'path' | 'parentId'>,
    map: Map<number, MenuDataItem>,
    visited = new Set<number>()
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

    if (item.id && visited.has(item.id)) {
      return `/${currentPath.replace(/^\//, '')}`
    }
    if (item.id) visited.add(item.id)

    const parent = map.get(item.parentId)
    if (!parent) {
      return `/${currentPath.replace(/^\//, '')}`
    }

    const parentPath = resolveFullMenuPath(parent, map, visited).replace(/\/$/, '')
    const childPath = currentPath.replace(/^\//, '')
    return parentPath ? `${parentPath}/${childPath}` : `/${childPath}`
  }

  function apiToTree(list: ApiDataItem[]): TreeNode[] {
    const roots: TreeNode[] = []
    const categories = new Map<string, TreeNode>()
    const menuTitles = new Map<number, string>()

    const collectMenuTitles = (nodes: TreeNode[]) => {
      nodes.forEach((node) => {
        if (node.menuId) menuTitles.set(node.menuId, node.title)
        if (node.children?.length) collectMenuTitles(node.children)
      })
    }
    collectMenuTitles(menuTree.value)

    list.forEach((item) => {
      const segments = (item.group || '未分类').split('/').filter(Boolean)
      let siblings = roots
      let categoryPath = ''
      segments.forEach((segment) => {
        categoryPath = categoryPath ? `${categoryPath}/${segment}` : segment
        let category = categories.get(categoryPath)
        if (!category) {
          category = {
            key: `api-category:${categoryPath}`,
            title: segment,
            kind: 'category',
            children: []
          }
          categories.set(categoryPath, category)
          siblings.push(category)
        }
        siblings = category.children || []
      })

      siblings.push({
        key: apiPermissionKey(item),
        title: item.name || item.path,
        kind: 'api',
        method: item.method,
        linkedMenuNames: (item.menuIds || []).map(
          (menuID) => menuTitles.get(menuID) || `菜单 #${menuID}`
        )
      })
    })

    return roots
  }

  function apiPermissionKey(item: Pick<ApiDataItem, 'path' | 'method'>) {
    return `api:${item.path},${item.method}`
  }

  function buildMenuApiIndex(apis: ApiDataItem[]) {
    menuApiKeys.clear()
    apis.forEach((api) => {
      const key = apiPermissionKey(api)
      ;(api.menuIds || []).forEach((menuID) => {
        const keys = menuApiKeys.get(menuID) || new Set<string>()
        keys.add(key)
        menuApiKeys.set(menuID, keys)
      })
    })
  }

  async function loadAllApis() {
    const pageSize = 500
    let page = 1
    let result: ApiDataItem[] = []
    let total = 0

    do {
      const data = await getAdminApiApi({ page, pageSize })
      const items = normalizeList(data)
      result = result.concat(items)
      total = Number(Array.isArray(data) ? result.length : (data.total ?? result.length))
      page += 1
      if (!items.length) break
    } while (result.length < total)

    return result
  }

  async function openPermission(row: RoleDataItem) {
    const requestId = ++permissionRequestId
    permissionRole.value = row
    permissionVisible.value = true
    permissionLoading.value = true
    lastAutoLinkedCount.value = 0
    autoCheckedApiKeys.clear()
    lockedApiKeys.clear()

    try {
      const [menus, rolePerms, apis] = await Promise.all([
        getAdminMenusApi({}),
        getRolePermissionsApi({ role: row.sid }),
        loadAllApis()
      ])
      if (requestId !== permissionRequestId || !permissionVisible.value) return
      const apiList = apis.map((item) => ({ ...item, menuIds: item.menuIds || [] }))
      buildMenuApiIndex(apiList)
      menuTree.value = menuToTree(normalizeList(menus))
      apiTree.value = apiToTree(apiList)

      await nextTick()
      const checked = normalizeList(rolePerms).map(String)
      setTreeCheckedKeys(
        menuTreeRef.value,
        checked.filter((item) => item.startsWith('menu:'))
      )
      apiTreeRef.value?.setCheckedKeys(checked.filter((item) => item.startsWith('api:')))
      previousCheckedMenuIDs = getCheckedMenuIDs()
      const required = requiredApiKeys(previousCheckedMenuIDs)
      const checkedApiKeys = getCheckedApiKeys()
      required.forEach((key) => {
        checkedApiKeys.add(key)
        autoCheckedApiKeys.add(key)
      })
      apiTreeRef.value?.setCheckedKeys([...checkedApiKeys])
      updateApiLocks(required)
      updatePermissionCounts()
    } catch (error) {
      if (requestId === permissionRequestId) throw error
    } finally {
      if (requestId === permissionRequestId) permissionLoading.value = false
    }
  }

  function closePermission() {
    permissionRequestId += 1
    permissionLoading.value = false
    permissionRole.value = {}
    menuTree.value = []
    apiTree.value = []
    menuApiKeys.clear()
    autoCheckedApiKeys.clear()
    lockedApiKeys.clear()
    previousCheckedMenuIDs = new Set()
    selectedMenuCount.value = 0
    selectedApiCount.value = 0
    lastAutoLinkedCount.value = 0
  }

  function setTreeCheckedKeys(tree: InstanceType<typeof ElTree> | undefined, keys: string[]) {
    if (!tree) return
    tree.setCheckedKeys([])
    keys.forEach((key) => tree.setChecked(key, true, false))
  }

  function getCheckedMenuIDs() {
    const nodes = (menuTreeRef.value?.getCheckedNodes(false, false) || []) as TreeNode[]
    return new Set(nodes.flatMap((node) => (node.menuId ? [node.menuId] : [])))
  }

  function getCheckedApiKeys() {
    return new Set(
      (apiTreeRef.value?.getCheckedKeys(false) || [])
        .map(String)
        .filter((key) => key.startsWith('api:'))
    )
  }

  function requiredApiKeys(menuIDs: Set<number>) {
    const result = new Set<string>()
    menuIDs.forEach((menuID) => menuApiKeys.get(menuID)?.forEach((key) => result.add(key)))
    return result
  }

  function updateApiLocks(required: Set<string>) {
    lockedApiKeys.clear()
    required.forEach((key) => lockedApiKeys.add(key))
    const updateNodes = (nodes: TreeNode[]) => {
      nodes.forEach((node) => {
        node.disabled = node.kind === 'api' && lockedApiKeys.has(node.key)
        if (node.children?.length) updateNodes(node.children)
      })
    }
    updateNodes(apiTree.value)
  }

  function handleMenuCheck() {
    const currentMenuIDs = getCheckedMenuIDs()
    const addedMenuIDs = new Set(
      [...currentMenuIDs].filter((menuID) => !previousCheckedMenuIDs.has(menuID))
    )
    const removedMenuIDs = new Set(
      [...previousCheckedMenuIDs].filter((menuID) => !currentMenuIDs.has(menuID))
    )
    const currentApiKeys = getCheckedApiKeys()
    const nextApiKeys = new Set(currentApiKeys)
    const stillRequired = requiredApiKeys(currentMenuIDs)
    let addedCount = 0

    requiredApiKeys(addedMenuIDs).forEach((key) => {
      if (!nextApiKeys.has(key)) {
        nextApiKeys.add(key)
        autoCheckedApiKeys.add(key)
        addedCount += 1
      }
    })

    if (removedMenuIDs.size) {
      requiredApiKeys(removedMenuIDs).forEach((key) => {
        if (autoCheckedApiKeys.has(key) && !stillRequired.has(key)) {
          nextApiKeys.delete(key)
          autoCheckedApiKeys.delete(key)
        }
      })
    }

    updateApiLocks(new Set())
    apiTreeRef.value?.setCheckedKeys([...nextApiKeys])
    updateApiLocks(stillRequired)
    previousCheckedMenuIDs = currentMenuIDs
    lastAutoLinkedCount.value = addedCount
    updatePermissionCounts()
  }

  function handleApiCheck() {
    const checked = getCheckedApiKeys()
    let restoredRequired = false
    lockedApiKeys.forEach((key) => {
      if (!checked.has(key)) {
        checked.add(key)
        restoredRequired = true
      }
    })
    if (restoredRequired) apiTreeRef.value?.setCheckedKeys([...checked])
    ;[...autoCheckedApiKeys].forEach((key) => {
      if (!checked.has(key)) autoCheckedApiKeys.delete(key)
    })
    lastAutoLinkedCount.value = 0
    updatePermissionCounts()
  }

  function updatePermissionCounts() {
    selectedMenuCount.value = getCheckedMenuIDs().size
    selectedApiCount.value = getCheckedApiKeys().size
  }

  function methodTagType(method: string) {
    const types: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'primary'> = {
      GET: 'success',
      POST: 'primary',
      PUT: 'warning',
      PATCH: 'warning',
      DELETE: 'danger'
    }
    return types[method] || 'info'
  }

  async function submitPermission() {
    submitting.value = true

    try {
      const list = [
        ...(apiTreeRef.value?.getCheckedKeys(false) || []),
        ...(menuTreeRef.value?.getCheckedKeys(false) || [])
      ]
        .map(String)
        .filter((key) => key.startsWith('api:') || key.startsWith('menu:'))

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

  .permission-editor {
    min-height: 420px;
  }

  .permission-summary {
    display: flex;
    gap: 10px;
    align-items: center;
    min-height: 36px;
    padding-inline: 4px;
    font-size: 13px;
    color: var(--art-gray-600);

    &__dot {
      width: 3px;
      height: 3px;
      background: var(--art-gray-400);
      border-radius: 50%;
    }

    &__linked {
      margin-inline-start: auto;
      color: var(--el-color-success);
    }
  }

  .permission-grid {
    display: grid;
    grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
    gap: 12px;
  }

  .permission-panel {
    min-width: 0;
    overflow: hidden;
    border: 1px solid var(--art-gray-300);
    border-radius: 10px;

    &__header {
      display: flex;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
      padding: 14px 16px;
      background: var(--art-gray-100);
      border-block-end: 1px solid var(--art-gray-300);

      h3,
      p {
        margin: 0;
      }

      h3 {
        font-size: 14px;
        font-weight: 600;
        color: var(--art-gray-900);
      }

      p {
        margin-top: 3px;
        font-size: 12px;
        color: var(--art-gray-500);
      }
    }

    :deep(.el-scrollbar__wrap) {
      padding: 10px 8px 14px;
    }
  }

  .tree-node {
    display: flex;
    flex: 1;
    gap: 8px;
    align-items: center;
    min-width: 0;
    padding-inline-end: 8px;

    &__title {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    &__link {
      flex: none;
      font-size: 14px;
      color: var(--el-color-success);
    }
  }

  @media (width <= 900px) {
    .permission-grid {
      grid-template-columns: 1fr;
    }

    .permission-summary__linked {
      display: none;
    }
  }
</style>
