<script>
  import { onMount, getContext } from 'svelte'
  import { api } from '../api/index.js'

  const { showAlert, showConfirm } = getContext('dialog')

  let courses = []
  let loading = true
  let showModal = false
  let editingCourse = null
  let formData = { title: '', date: '', time_slot: '', capacity: 15 }
  let formError = ''
  let formLoading = false
  let showBookingsModal = false
  let selectedCourseBookings = []
  let selectedCourseTitle = ''
  let showWaitlistModal = false
  let selectedCourseWaitlist = []
  let selectedCourseWaitlistCount = 0

  onMount(async () => {
    await loadCourses()
  })

  async function loadCourses() {
    loading = true
    try {
      const res = await api.getCourses()
      courses = res.data || []
    } catch (e) {
      console.error(e)
    }
    loading = false
  }

  function openCreate() {
    editingCourse = null
    formData = { title: '', date: '', time_slot: '', capacity: 15 }
    formError = ''
    showModal = true
  }

  function openEdit(course) {
    editingCourse = course
    formData = {
      title: course.title,
      date: course.date,
      time_slot: course.time_slot,
      capacity: course.capacity,
    }
    formError = ''
    showModal = true
  }

  function closeModal() {
    showModal = false
    editingCourse = null
  }

  async function submitForm() {
    if (!formData.title || !formData.date || !formData.time_slot || !formData.capacity) {
      formError = '请填写所有必填字段'
      return
    }
    if (formData.capacity < 1) {
      formError = '名额上限至少为1'
      return
    }
    formLoading = true
    formError = ''
    try {
      if (editingCourse) {
        await api.updateCourse(editingCourse.id, formData)
        await showAlert('更新成功', '课程信息已更新')
      } else {
        await api.createCourse(formData)
        await showAlert('发布成功', '课程已成功发布')
      }
      closeModal()
      await loadCourses()
    } catch (e) {
      formError = e.message
    }
    formLoading = false
  }

  async function deleteCourse(course) {
    const confirmed = await showConfirm(
      '确认删除',
      `确认删除课程「${course.title}」（${course.date} ${course.time_slot}）吗？\n有预约的课程无法删除。`,
      '确认删除',
      '取消'
    )
    if (!confirmed) return

    try {
      await api.deleteCourse(course.id)
      await showAlert('删除成功', '课程已删除')
      await loadCourses()
    } catch (e) {
      await showAlert('删除失败', e.message)
    }
  }

  async function viewBookings(course) {
    selectedCourseTitle = course.title
    try {
      const res = await api.getCourseBookings(course.id)
      selectedCourseBookings = res.data || []
      showBookingsModal = true
    } catch (e) {
      await showAlert('查询失败', e.message)
    }
  }

  function closeBookings() {
    showBookingsModal = false
  }

  async function viewWaitlist(course) {
    selectedCourseTitle = course.title
    try {
      const res = await api.getCourseWaitlist(course.id)
      selectedCourseWaitlist = res.data || []
      selectedCourseWaitlistCount = res.count || 0
      showWaitlistModal = true
    } catch (e) {
      await showAlert('查询失败', e.message)
    }
  }

  function closeWaitlist() {
    showWaitlistModal = false
  }

  async function markAttended(booking) {
    const confirmed = await showConfirm(
      '标记到课',
      `确认将 ${booking.user_name}（${booking.user_phone}）标记为已到课？`,
      '确认标记',
      '取消'
    )
    if (!confirmed) return

    try {
      await api.markAttendance(booking.id)
      await loadCourses()
      const res = await api.getCourseBookings(booking.course_id)
      selectedCourseBookings = res.data || []
      await showAlert('操作成功', '已标记为到课')
    } catch (e) {
      await showAlert('操作失败', e.message)
    }
  }
</script>

<div class="page">
  <div class="page-header">
    <h2>⚙️ 课程管理</h2>
    <button class="btn btn-primary" on:click={openCreate}>+ 发布新课程</button>
  </div>

  {#if loading}
    <div class="loading">加载中...</div>
  {:else if courses.length === 0}
    <div class="empty">暂无课程，点击上方按钮发布</div>
  {:else}
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>课程名称</th>
            <th>日期</th>
            <th>时段</th>
            <th>名额</th>
            <th>已预约</th>
            <th>候补</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {#each courses as course}
            <tr>
              <td><strong>{course.title}</strong></td>
              <td>{course.date}</td>
              <td>{course.time_slot}</td>
              <td>{course.capacity}</td>
              <td>
                <span class:full={course.booked >= course.capacity}>
                  {course.booked}/{course.capacity}
                </span>
              </td>
              <td>
                <span class:waitlist-full={(course.waitlist_count || 0) > 0}>
                  {course.waitlist_count || 0}人
                </span>
              </td>
              <td>
                {#if course.booked >= course.capacity}
                  <span class="badge badge-full">已满</span>
                {:else}
                  <span class="badge badge-open">报名中</span>
                {/if}
              </td>
              <td class="actions">
                <button class="btn-sm btn-info" on:click={() => viewBookings(course)}>名单</button>
                <button class="btn-sm btn-warn" on:click={() => viewWaitlist(course)}>候补</button>
                <button class="btn-sm btn-warn" on:click={() => openEdit(course)}>编辑</button>
                <button class="btn-sm btn-del" on:click={() => deleteCourse(course)}>删除</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if showModal}
  <div class="modal-overlay" on:click={closeModal} on:keydown={(e) => e.key === 'Escape' && closeModal()} role="dialog" aria-modal="true">
    <div class="modal" on:click|stopPropagation role="document">
      <div class="modal-header">
        <h3>{editingCourse ? '编辑课程' : '发布新课程'}</h3>
        <button class="modal-close" on:click={closeModal} aria-label="关闭">✕</button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label for="course-title">课程名称 *</label>
          <input id="course-title" type="text" bind:value={formData.title} placeholder="如：剪纸入门体验课" />
        </div>
        <div class="form-group">
          <label for="course-date">日期 *</label>
          <input id="course-date" type="date" bind:value={formData.date} />
        </div>
        <div class="form-group">
          <label for="course-slot">时段 *</label>
          <input id="course-slot" type="text" bind:value={formData.time_slot} placeholder="如：09:00-11:00" />
        </div>
        <div class="form-group">
          <label for="course-capacity">名额上限 *</label>
          <input id="course-capacity" type="number" bind:value={formData.capacity} min="1" />
        </div>
        {#if formError}
          <div class="msg-error">{formError}</div>
        {/if}
        <button class="btn btn-primary btn-block" on:click={submitForm} disabled={formLoading}>
          {formLoading ? '提交中...' : (editingCourse ? '保存修改' : '发布课程')}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if showBookingsModal}
  <div class="modal-overlay" on:click={closeBookings} on:keydown={(e) => e.key === 'Escape' && closeBookings()} role="dialog" aria-modal="true">
    <div class="modal modal-lg" on:click|stopPropagation role="document">
      <div class="modal-header">
        <h3>预约名单 - {selectedCourseTitle}</h3>
        <button class="modal-close" on:click={closeBookings} aria-label="关闭">✕</button>
      </div>
      <div class="modal-body">
        {#if selectedCourseBookings.length === 0}
          <div class="empty-sm">暂无预约</div>
        {:else}
          <table class="inner-table">
            <thead>
              <tr>
                <th>序号</th>
                <th>姓名</th>
                <th>手机号</th>
                <th>状态</th>
                <th>预约时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {#each selectedCourseBookings as booking, i}
                <tr>
                  <td>{i + 1}</td>
                  <td>{booking.user_name}</td>
                  <td>{booking.user_phone}</td>
                  <td>
                    {#if booking.status === 'booked'}
                      <span class="status-badge status-booked">已预约</span>
                    {:else if booking.status === 'attended'}
                      <span class="status-badge status-attended">已到课</span>
                    {:else}
                      <span class="status-badge status-cancelled">已取消</span>
                    {/if}
                  </td>
                  <td>{new Date(booking.created_at).toLocaleString('zh-CN')}</td>
                  <td>
                    {#if booking.status === 'booked'}
                      <button class="btn-sm btn-info" on:click={() => markAttended(booking)}>到课</button>
                    {:else}
                      -
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    </div>
  </div>
{/if}

{#if showWaitlistModal}
  <div class="modal-overlay" on:click={closeWaitlist} on:keydown={(e) => e.key === 'Escape' && closeWaitlist()} role="dialog" aria-modal="true">
    <div class="modal modal-lg" on:click|stopPropagation role="document">
      <div class="modal-header">
        <h3>候补队列 - {selectedCourseTitle} ({selectedCourseWaitlistCount}人)</h3>
        <button class="modal-close" on:click={closeWaitlist} aria-label="关闭">✕</button>
      </div>
      <div class="modal-body">
        {#if selectedCourseWaitlist.length === 0}
          <div class="empty-sm">暂无候补人员</div>
        {:else}
          <table class="inner-table">
            <thead>
              <tr>
                <th>顺位</th>
                <th>姓名</th>
                <th>手机号</th>
                <th>状态</th>
                <th>候补时间</th>
                <th>到期时间</th>
              </tr>
            </thead>
            <tbody>
              {#each selectedCourseWaitlist as item, i}
                <tr>
                  <td><strong>{item.position}</strong></td>
                  <td>{item.user_name}</td>
                  <td>{item.user_phone}</td>
                  <td>
                    {#if item.status === 'notified'}
                      <span class="status-badge status-notified">待确认</span>
                    {:else}
                      <span class="status-badge status-waiting">排队中</span>
                    {/if}
                  </td>
                  <td>{new Date(item.created_at).toLocaleString('zh-CN')}</td>
                  <td>
                    {#if item.expires_at}
                      <span class="expire-warning">{new Date(item.expires_at).toLocaleString('zh-CN')}</span>
                    {:else}
                      -
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .page { max-width: 1200px; margin: 0 auto; }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .page-header h2 { font-size: 1.5rem; }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary { background: #8b2500; color: white; }
  .btn-primary:hover { background: #a0522d; }
  .btn-primary:disabled { background: #ccc; }
  .btn-block { width: 100%; }

  .btn-sm {
    padding: 4px 10px;
    border: none;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-info { background: #cce5ff; color: #004085; }
  .btn-info:hover { background: #b8daff; }
  .btn-warn { background: #fff3cd; color: #856404; }
  .btn-warn:hover { background: #ffeaa7; }
  .btn-del { background: #f8d7da; color: #721c24; }
  .btn-del:hover { background: #f5c6cb; }

  .loading, .empty {
    text-align: center;
    padding: 60px;
    color: #888;
  }

  .empty-sm {
    text-align: center;
    padding: 30px;
    color: #888;
  }

  .table-wrap {
    background: white;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th {
    background: #faf5f0;
    padding: 12px 16px;
    text-align: left;
    font-size: 0.85rem;
    color: #8b7355;
    font-weight: 600;
  }

  td {
    padding: 12px 16px;
    border-top: 1px solid #f0e8e0;
    font-size: 0.9rem;
  }

  .actions {
    display: flex;
    gap: 6px;
  }

  .full { color: #dc3545; font-weight: 600; }

  .badge {
    padding: 3px 8px;
    border-radius: 10px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .badge-open { background: #d4edda; color: #155724; }
  .badge-full { background: #f8d7da; color: #721c24; }

  .status-badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 10px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .status-booked { background: #d4edda; color: #155724; }
  .status-attended { background: #cce5ff; color: #004085; }
  .status-cancelled { background: #f8d7da; color: #721c24; }
  .status-waiting { background: #ffe5cc; color: #a04000; }
  .status-notified { background: #ffc107; color: #523e02; }

  .waitlist-full {
    color: #fd7e14;
    font-weight: 600;
  }

  .expire-warning {
    color: #dc3545;
    font-weight: 600;
  }

  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .modal {
    background: white;
    border-radius: 16px;
    width: 480px;
    max-width: 90vw;
    box-shadow: 0 8px 32px rgba(0,0,0,0.2);
    animation: slideUp 0.25s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-lg { width: 720px; }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #f0e8e0;
  }

  .modal-header h3 { font-size: 1.1rem; }

  .modal-close {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    color: #888;
  }

  .modal-body { padding: 24px; }

  .form-group { margin-bottom: 16px; }

  .form-group label {
    display: block;
    margin-bottom: 6px;
    font-size: 0.85rem;
    color: #8b7355;
  }

  .form-group input {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #ddd;
    border-radius: 8px;
    font-size: 0.9rem;
  }

  .form-group input:focus {
    outline: none;
    border-color: #8b2500;
  }

  .msg-error {
    background: #f8d7da;
    color: #721c24;
    padding: 10px;
    border-radius: 8px;
    font-size: 0.85rem;
    margin-bottom: 12px;
    text-align: center;
  }

  .inner-table {
    width: 100%;
    border-collapse: collapse;
  }

  .inner-table th {
    background: #faf5f0;
    padding: 8px 12px;
    text-align: left;
    font-size: 0.8rem;
    color: #8b7355;
  }

  .inner-table td {
    padding: 8px 12px;
    border-top: 1px solid #f0e8e0;
    font-size: 0.85rem;
  }
</style>
