<template>
  <div class="admin-page art-full-height">
    <ElCard class="search-card" shadow="never">
      <ElForm :model="searchForm" label-width="82px">
        <ElRow :gutter="16">
          <ElCol :xs="24" :md="8">
            <ElFormItem label="用户ID"><ElInput v-model="searchForm.id" clearable /></ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8">
            <ElFormItem label="用户名"
              ><ElInput v-model="searchForm.username" clearable
            /></ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8">
            <ElFormItem label="昵称"
              ><ElInput v-model="searchForm.nickname" clearable
            /></ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8">
            <ElFormItem label="邮箱"><ElInput v-model="searchForm.email" clearable /></ElFormItem>
          </ElCol>
          <ElCol :xs="24" :md="8">
            <ElFormItem label="手机号"><ElInput v-model="searchForm.phone" clearable /></ElFormItem>
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
          <span>用户列表</span>
          <ElSpace>
            <ElButton type="primary" @click="openCreate">新增用户</ElButton>
            <ElButton :loading="loading" @click="loadData">刷新</ElButton>
          </ElSpace>
        </div>
      </template>

      <ElTable v-loading="loading" :data="rows" height="100%" border stripe>
        <ElTableColumn prop="id" label="ID" width="90" />
        <ElTableColumn prop="username" label="用户名" min-width="140" />
        <ElTableColumn prop="nickname" label="昵称" min-width="140" />
        <ElTableColumn prop="email" label="邮箱" min-width="180" />
        <ElTableColumn prop="phone" label="手机号" min-width="150" />
        <ElTableColumn label="角色" min-width="180">
          <template #default="{ row }">
            <ElTag v-for="role in row.roles || []" :key="role" class="mr-1">
              {{ roleMap[role] || role }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="170" fixed="right">
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

    <ElDialog
      v-model="dialogVisible"
      :title="form.id ? '编辑用户' : '新增用户'"
      width="520px"
      align-center
      destroy-on-close
      @closed="clearFormValidate"
    >
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
        <ElFormItem label="用户名" prop="username">
          <ElInput v-model="form.username" />
        </ElFormItem>
        <ElFormItem label="密码" :prop="form.id ? undefined : 'password'">
          <ElInput v-model="form.password" type="password" show-password />
        </ElFormItem>
        <ElFormItem label="昵称">
          <ElInput v-model="form.nickname" />
        </ElFormItem>
        <ElFormItem label="邮箱">
          <ElInput v-model="form.email" />
        </ElFormItem>
        <ElFormItem label="手机号">
          <ElInput v-model="form.phone" />
        </ElFormItem>
        <ElFormItem label="角色" prop="roles">
          <ElSelect v-model="form.roles" multiple class="w-full">
            <ElOption v-for="item in roles" :key="item.sid" :label="item.name" :value="item.sid" />
          </ElSelect>
        </ElFormItem>
      </ElForm>

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
    createAdminUserApi,
    deleteAdminUserApi,
    getAdminUsersApi,
    getRolesApi,
    updateAdminUserApi
  } from '@/api/admin/permission'
  import type {
    AdminUserCreateRequest,
    AdminUserDataItem,
    AdminUserUpdateRequest,
    GetAdminUsersRequest,
    RoleDataItem
  } from '@/api/admin/types'

  defineOptions({ name: 'AdminUser' })

  interface UserForm {
    id?: number
    username: string
    password: string
    nickname: string
    email: string
    phone: string
    roles: string[]
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
  const rows = ref<AdminUserDataItem[]>([])
  const roles = ref<RoleDataItem[]>([])
  const roleMap = ref<Record<string, string>>({})
  const dialogVisible = ref(false)
  const formRef = ref<FormInstance>()
  const pagination = reactive({ current: 1, pageSize: 20, total: 0 })
  const searchForm = reactive<Partial<GetAdminUsersRequest>>({
    id: '',
    username: '',
    nickname: '',
    email: '',
    phone: ''
  })
  const form = reactive<UserForm>(createDefaultForm())
  const rules: FormRules<UserForm> = {
    username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请设置密码', trigger: 'blur' }],
    roles: [{ required: true, message: '请分配角色', trigger: 'change' }]
  }

  function createDefaultForm(): UserForm {
    return {
      id: undefined,
      username: '',
      password: '',
      nickname: '',
      email: '',
      phone: '',
      roles: []
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

  function resetForm(data: Partial<UserForm> = {}) {
    Object.assign(form, createDefaultForm(), data)
    nextTick(clearFormValidate)
  }

  function clearFormValidate() {
    formRef.value?.clearValidate()
  }

  async function loadRoles() {
    const data = await getRolesApi({ page: 1, pageSize: 999 })
    const list = normalizeList(data)
    roles.value = list
    roleMap.value = Object.fromEntries(list.map((item) => [item.sid, item.name]))
  }

  async function loadData() {
    loading.value = true

    try {
      await loadRoles()
      const data = await getAdminUsersApi({
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
    Object.assign(searchForm, { id: '', username: '', nickname: '', email: '', phone: '' })
    pagination.current = 1
    await loadData()
  }

  function openCreate() {
    resetForm()
    dialogVisible.value = true
  }

  function openEdit(row: AdminUserDataItem) {
    resetForm({ ...row, password: '' })
    dialogVisible.value = true
  }

  async function remove(row: AdminUserDataItem) {
    await ElMessageBox.confirm(`确定删除用户「${row.username || row.id}」吗？`, '删除确认', {
      type: 'warning'
    })
    await deleteAdminUserApi({ id: row.id })
    ElMessage.success('删除成功')
    await loadData()
  }

  function createBasePayload() {
    return {
      username: form.username,
      nickname: form.nickname,
      email: form.email,
      phone: form.phone,
      roles: [...form.roles]
    }
  }

  async function submit() {
    if (!formRef.value) {
      throw new Error('用户表单未初始化')
    }

    await formRef.value.validate()
    submitting.value = true

    try {
      const basePayload = createBasePayload()
      if (form.id) {
        const payload: AdminUserUpdateRequest = { ...basePayload, id: form.id }
        if (form.password) {
          payload.password = form.password
        }
        await updateAdminUserApi(payload)
      } else {
        const payload: AdminUserCreateRequest = { ...basePayload, password: form.password }
        await createAdminUserApi(payload)
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

  .mr-1 {
    margin-right: 4px;
  }

  .w-full {
    width: 100%;
  }
</style>
