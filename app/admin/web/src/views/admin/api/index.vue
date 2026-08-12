<template>
  <div class="admin-page art-full-height">
    <ElCard shadow="never" class="search-card">
      <ElForm :model="searchForm" label-width="82px">
        <ElRow :gutter="16">
          <ElCol :xs="24" :md="8">
            <ElFormItem label="API 分类">
              <ElInput v-model="searchForm.group" clearable placeholder="支持查询任意层级" />
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8">
            <ElFormItem label="接口名称">
              <ElInput v-model="searchForm.name" clearable />
            </ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8" class="search-actions">
            <ElButton type="primary" :loading="loading" @click="search">查询</ElButton>
            <ElButton @click="resetSearch">重置</ElButton>
          </ElCol>
        </ElRow>
      </ElForm>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <div class="card-header__title">接口管理</div>
            <div class="card-header__hint">分类使用“/”分隔层级，接口可关联一个或多个菜单</div>
          </div>
          <ElSpace>
            <ElButton type="primary" @click="openCreate">新增接口</ElButton>
            <ElButton :loading="loading" @click="loadData">刷新</ElButton>
          </ElSpace>
        </div>
      </template>

      <ElTable v-loading="loading" :data="rows" height="100%" border stripe>
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn label="分类" min-width="190">
          <template #default="{ row }">
            <div class="category-path" :title="row.group">
              <template v-for="(part, index) in splitGroup(row.group)" :key="`${part}-${index}`">
                <span v-if="index" class="category-path__separator">/</span>
                <span>{{ part }}</span>
              </template>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="name" label="名称" min-width="170" show-overflow-tooltip />
        <ElTableColumn prop="method" label="方法" width="100">
          <template #default="{ row }">
            <ElTag :type="methodTagType(row.method)" effect="plain">{{ row.method }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="path" label="路径" min-width="220" show-overflow-tooltip />
        <ElTableColumn label="关联菜单" min-width="180">
          <template #default="{ row }">
            <ElTooltip
              v-if="row.menuIds?.length"
              :content="relatedMenuNames(row.menuIds).join('、')"
              placement="top"
            >
              <ElTag type="success" effect="plain">已关联 {{ row.menuIds.length }} 项</ElTag>
            </ElTooltip>
            <span v-else class="muted-text">未关联</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="150" fixed="right">
          <template #default="{ row }">
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

    <ElDrawer v-model="drawerVisible" :title="form.id ? '编辑接口' : '新增接口'" size="500px">
      <ElForm ref="formRef" :model="form" :rules="rules" label-position="top">
        <ElFormItem label="API 分类" prop="group">
          <ElAutocomplete
            v-model="form.group"
            :fetch-suggestions="suggestGroups"
            class="w-full"
            clearable
            placeholder="如：权限管理/角色"
          />
          <div class="field-hint">使用“/”创建上下级分类，可直接输入新的分类路径</div>
        </ElFormItem>

        <ElFormItem label="接口名称" prop="name">
          <ElInput v-model="form.name" maxlength="100" show-word-limit />
        </ElFormItem>

        <div class="form-grid">
          <ElFormItem label="请求方法" prop="method">
            <ElSelect v-model="form.method" class="w-full">
              <ElOption v-for="item in methods" :key="item" :label="item" :value="item" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="接口路径" prop="path">
            <ElInput v-model="form.path" placeholder="/v1/example" />
          </ElFormItem>
        </div>

        <ElFormItem label="关联菜单">
          <ElTreeSelect
            v-model="form.menuIds"
            :data="menuOptions"
            node-key="id"
            :props="menuTreeProps"
            multiple
            show-checkbox
            check-strictly
            clearable
            collapse-tags
            collapse-tags-tooltip
            class="w-full"
            placeholder="选择使用此接口的菜单"
          />
          <div class="field-hint">配置后，在角色权限中勾选这些菜单会自动选中该接口</div>
        </ElFormItem>
      </ElForm>

      <template #footer>
        <ElButton @click="drawerVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="submit">保存</ElButton>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
  import {
    createAdminApiApi,
    deleteAdminApiApi,
    getAdminApiApi,
    getAdminMenusApi,
    updateAdminApiApi
  } from '@/api/admin/permission'
  import type {
    ApiCreateRequest,
    ApiDataItem,
    GetApisRequest,
    MenuDataItem
  } from '@/api/admin/types'
  import { formatMenuTitle } from '@/utils/router'

  defineOptions({ name: 'AdminApi' })

  type ApiForm = ApiCreateRequest & { id?: number }
  type MenuOption = MenuDataItem & { displayTitle: string; children: MenuOption[] }
  type ListPayload<T> = T[] | { list?: T[]; records?: T[]; items?: T[]; total?: number }

  const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']
  const loading = ref(false)
  const submitting = ref(false)
  const rows = ref<ApiDataItem[]>([])
  const menuOptions = ref<MenuOption[]>([])
  const menuNameMap = ref(new Map<number, string>())
  const groupOptions = ref<string[]>([])
  const drawerVisible = ref(false)
  const formRef = ref<FormInstance>()
  const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
  const searchForm = reactive<Partial<GetApisRequest>>({ group: '', name: '' })
  const form = reactive<ApiForm>(createDefaultForm())
  const menuTreeProps = { label: 'displayTitle', children: 'children' } as const
  const rules: FormRules<ApiForm> = {
    group: [
      { required: true, message: '请输入 API 分类', trigger: 'blur' },
      {
        pattern: /^(?!\/)(?!.*\/\s*\/)(?!.*\/$).+$/,
        message: '请使用“/”分隔分类，不能以“/”开头或结尾',
        trigger: 'blur'
      }
    ],
    name: [{ required: true, message: '请输入接口名称', trigger: 'blur' }],
    method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
    path: [
      { required: true, message: '请输入接口路径', trigger: 'blur' },
      { pattern: /^\//, message: '接口路径需要以“/”开头', trigger: 'blur' }
    ]
  }

  function createDefaultForm(): ApiForm {
    return { id: undefined, group: '', name: '', method: 'GET', path: '', menuIds: [] }
  }

  function normalizeList<T>(data: ListPayload<T>): T[] {
    if (Array.isArray(data)) return data
    return data.list || data.records || data.items || []
  }

  function toMenuTree(list: MenuDataItem[]): MenuOption[] {
    const map = new Map<number, MenuOption>()
    const result: MenuOption[] = []
    menuNameMap.value = new Map()

    list.forEach((item) => {
      if (!item.id) return
      const displayTitle = formatMenuTitle(item.title || item.name || item.path)
      map.set(item.id, { ...item, displayTitle, children: [] })
      menuNameMap.value.set(item.id, displayTitle)
    })
    map.forEach((node) => {
      const parent = node.parentId ? map.get(node.parentId) : undefined
      if (parent) parent.children.push(node)
      else result.push(node)
    })
    return result
  }

  async function loadMenuOptions() {
    const data = await getAdminMenusApi({})
    menuOptions.value = toMenuTree(normalizeList(data))
  }

  async function loadData() {
    loading.value = true
    try {
      const data = await getAdminApiApi({
        ...searchForm,
        page: pagination.current,
        pageSize: pagination.pageSize
      })
      const list = normalizeList(data)
      rows.value = list.map((item) => ({ ...item, menuIds: item.menuIds || [] }))
      pagination.total = Number(Array.isArray(data) ? list.length : (data.total ?? list.length))
      groupOptions.value = Array.isArray(data) ? [] : [...new Set(data.groups || [])]
    } finally {
      loading.value = false
    }
  }

  async function search() {
    pagination.current = 1
    await loadData()
  }

  async function resetSearch() {
    searchForm.group = ''
    searchForm.name = ''
    await search()
  }

  function resetForm(data: Partial<ApiForm> = {}) {
    Object.assign(form, createDefaultForm(), data, { menuIds: [...(data.menuIds || [])] })
    nextTick(() => formRef.value?.clearValidate())
  }

  function openCreate() {
    resetForm()
    drawerVisible.value = true
  }

  function openEdit(row: ApiDataItem) {
    resetForm(row)
    drawerVisible.value = true
  }

  async function remove(row: ApiDataItem) {
    await ElMessageBox.confirm(`确定删除接口「${row.name || row.path}」吗？`, '删除确认', {
      type: 'warning'
    })
    await deleteAdminApiApi({ id: row.id })
    ElMessage.success('删除成功')
    await loadData()
  }

  async function submit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitting.value = true
    try {
      const payload = {
        ...form,
        group: form.group
          .split('/')
          .map((part) => part.trim())
          .join('/'),
        name: form.name.trim(),
        path: form.path.trim(),
        menuIds: [...new Set(form.menuIds)]
      }
      if (payload.id) await updateAdminApiApi({ ...payload, id: payload.id })
      else await createAdminApiApi(payload)
      ElMessage.success('保存成功')
      drawerVisible.value = false
      await loadData()
    } finally {
      submitting.value = false
    }
  }

  function suggestGroups(query: string, callback: (items: Array<{ value: string }>) => void) {
    const keyword = query.trim().toLowerCase()
    const items = groupOptions.value
      .filter((item) => !keyword || item.toLowerCase().includes(keyword))
      .map((value) => ({ value }))
    callback(items)
  }

  function splitGroup(group: string) {
    return (group || '未分类').split('/').filter(Boolean)
  }

  function relatedMenuNames(ids: number[]) {
    return ids.map((id) => menuNameMap.value.get(id) || `菜单 #${id}`)
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

  onMounted(async () => {
    await Promise.all([loadMenuOptions(), loadData()])
  })
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
    gap: 16px;
    align-items: center;
    justify-content: space-between;

    &__title {
      font-weight: 500;
    }

    &__hint {
      margin-top: 4px;
      font-size: 12px;
      color: var(--art-gray-600);
    }
  }

  .search-actions {
    display: flex;
    align-items: flex-start;
  }

  .category-path {
    display: flex;
    min-width: 0;
    overflow: hidden;
    color: var(--art-gray-800);
    text-overflow: ellipsis;
    white-space: nowrap;

    &__separator {
      margin-inline: 6px;
      color: var(--art-gray-400);
    }
  }

  .muted-text,
  .field-hint {
    font-size: 12px;
    color: var(--art-gray-500);
  }

  .field-hint {
    margin-top: 6px;
    line-height: 1.5;
  }

  .form-grid {
    display: grid;
    grid-template-columns: 140px minmax(0, 1fr);
    gap: 16px;
  }

  .pagination {
    justify-content: flex-end;
    margin-top: 16px;
  }

  .w-full {
    width: 100%;
  }

  @media (width <= 640px) {
    .card-header {
      align-items: flex-start;
    }

    .card-header__hint {
      display: none;
    }

    .form-grid {
      grid-template-columns: 1fr;
      gap: 0;
    }
  }
</style>
