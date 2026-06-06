<script>
  import { onMount, getContext } from 'svelte'
  import { api } from '../api/index.js'

  export let userPhone = ''
  export let userName = ''

  const { showAlert, showConfirm } = getContext('dialog')

  let courses = []
  let loading = true
  let bookingModal = false
  let selectedCourse = null
  let bookingPhone = ''
  let bookingName = ''
  let bookingMsg = ''
  let bookingError = ''
  let bookingLoading = false
  let phoneError = ''

  const phoneRegex = /^1[3-9]\d{9}$/

  function validatePhone(phone) {
    if (!phone.trim()) {
      phoneError = '请输入手机号'
      return false
    }
    if (!phoneRegex.test(phone.trim())) {
      phoneError = '请输入正确的11位手机号'
      return false
    }
    phoneError = ''
    return true
  }

  $: if (bookingPhone) validatePhone(bookingPhone)

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

  function openBooking(course) {
    selectedCourse = course
    bookingPhone = userPhone || ''
    bookingName = userName || ''
    bookingMsg = ''
    bookingError = ''
    phoneError = ''
    bookingModal = true
  }

  function closeBooking() {
    bookingModal = false
    selectedCourse = null
  }

  async function submitBooking() {
    if (!selectedCourse) return
    if (!bookingName.trim()) {
      bookingError = '请填写姓名'
      return
    }
    if (!validatePhone(bookingPhone)) {
      bookingError = phoneError
      return
    }
    bookingLoading = true
    bookingError = ''
    try {
      const res = await api.createBooking({
        course_id: selectedCourse.id,
        user_name: bookingName.trim(),
        user_phone: bookingPhone.trim(),
      })
      userPhone = bookingPhone.trim()
      userName = bookingName.trim()
      bookingMsg = `预约成功！\n${res.data.course_title}\n${res.data.date} ${res.data.time_slot}\n剩余名额：${res.data.remaining}`
      await loadCourses()
    } catch (e) {
      bookingError = e.message
    }
    bookingLoading = false
  }

  function getCapacityPercent(course) {
    if (!course.capacity) return 0
    return Math.min(100, Math.round((course.booked / course.capacity) * 100))
  }

  function getCapacityColor(course) {
    const pct = getCapacityPercent(course)
    if (pct >= 100) return '#dc3545'
    if (pct >= 80) return '#fd7e14'
    return '#28a745'
  }
</script>

<div class="page">
  <div class="page-header">
    <h2>📋 课程预约</h2>
    <button class="btn-refresh" on:click={loadCourses}>🔄 刷新</button>
  </div>

  {#if loading}
    <div class="loading">加载中...</div>
  {:else if courses.length === 0}
    <div class="empty">暂无可预约的课程</div>
  {:else}
    <div class="course-grid">
      {#each courses as course}
        <div class="course-card">
          <div class="card-header">
            <h3>{course.title}</h3>
            {#if course.booked >= course.capacity}
              <span class="badge badge-full">已满</span>
            {:else}
              <span class="badge badge-open">可预约</span>
            {/if}
          </div>
          <div class="card-body">
            <div class="info-row">
              <span class="info-label">📅 日期</span>
              <span>{course.date}</span>
            </div>
            <div class="info-row">
              <span class="info-label">🕐 时段</span>
              <span>{course.time_slot}</span>
            </div>
            <div class="info-row">
              <span class="info-label">👥 名额</span>
              <span>{course.booked}/{course.capacity}</span>
            </div>
            <div class="capacity-bar">
              <div class="capacity-fill" style="width: {getCapacityPercent(course)}%; background: {getCapacityColor(course)}"></div>
            </div>
          </div>
          <div class="card-footer">
            {#if course.booked >= course.capacity}
              <button class="btn btn-disabled" disabled>名额已满</button>
            {:else}
              <button class="btn btn-primary" on:click={() => openBooking(course)}>立即预约</button>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if bookingModal}
  <div class="modal-overlay" on:click={closeBooking} on:keydown={(e) => e.key === 'Escape' && closeBooking()} role="dialog" aria-modal="true">
    <div class="modal" on:click|stopPropagation role="document">
      <div class="modal-header">
        <h3>预约课程</h3>
        <button class="modal-close" on:click={closeBooking} aria-label="关闭">✕</button>
      </div>
      <div class="modal-body">
        {#if selectedCourse}
          <div class="booking-course-info">
            <p><strong>{selectedCourse.title}</strong></p>
            <p>📅 {selectedCourse.date} | 🕐 {selectedCourse.time_slot}</p>
            <p>剩余名额：{selectedCourse.capacity - selectedCourse.booked}</p>
          </div>
        {/if}

        {#if bookingMsg}
          <div class="msg-success">{bookingMsg}</div>
          <button class="btn btn-primary btn-block" style="margin-top: 16px;" on:click={closeBooking}>
            关闭
          </button>
        {:else}
          <div class="form-group">
            <label for="booking-name">姓名 *</label>
            <input
              id="booking-name"
              type="text"
              bind:value={bookingName}
              placeholder="请输入您的姓名"
            />
          </div>
          <div class="form-group">
            <label for="booking-phone">手机号 *</label>
            <input
              id="booking-phone"
              type="tel"
              bind:value={bookingPhone}
              placeholder="请输入11位手机号"
              maxlength="11"
            />
            {#if phoneError}
              <div class="field-error">{phoneError}</div>
            {/if}
          </div>
          {#if bookingError}
            <div class="msg-error">{bookingError}</div>
          {/if}
          <button class="btn btn-primary btn-block" on:click={submitBooking} disabled={bookingLoading}>
            {bookingLoading ? '提交中...' : '确认预约'}
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .page {
    max-width: 1200px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .page-header h2 {
    font-size: 1.5rem;
  }

  .btn-refresh {
    padding: 8px 16px;
    background: white;
    border: 1px solid #ddd;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.85rem;
  }

  .btn-refresh:hover { background: #f8f8f8; }

  .loading, .empty {
    text-align: center;
    padding: 60px;
    color: #888;
    font-size: 1.1rem;
  }

  .course-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 20px;
  }

  .course-card {
    background: white;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.08);
    overflow: hidden;
    transition: transform 0.2s, box-shadow 0.2s;
  }

  .course-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  }

  .card-header {
    padding: 16px 20px;
    border-bottom: 1px solid #f0e8e0;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .card-header h3 {
    font-size: 1.05rem;
    color: #2c1810;
  }

  .badge {
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .badge-open { background: #d4edda; color: #155724; }
  .badge-full { background: #f8d7da; color: #721c24; }

  .card-body {
    padding: 16px 20px;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    padding: 6px 0;
    font-size: 0.9rem;
  }

  .info-label {
    color: #8b7355;
  }

  .capacity-bar {
    height: 6px;
    background: #f0e8e0;
    border-radius: 3px;
    margin-top: 12px;
    overflow: hidden;
  }

  .capacity-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s;
  }

  .card-footer {
    padding: 12px 20px;
    border-top: 1px solid #f0e8e0;
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #8b2500;
    color: white;
    width: 100%;
  }

  .btn-primary:hover { background: #a0522d; }

  .btn-disabled {
    background: #ddd;
    color: #999;
    width: 100%;
    cursor: not-allowed;
  }

  .btn-block { width: 100%; }

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
    width: 420px;
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

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #f0e8e0;
  }

  .modal-header h3 { font-size: 1.15rem; }

  .modal-close {
    background: none;
    border: none;
    font-size: 1.2rem;
    cursor: pointer;
    color: #888;
  }

  .modal-body { padding: 24px; }

  .booking-course-info {
    background: #faf5f0;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 20px;
    line-height: 1.8;
  }

  .form-group {
    margin-bottom: 16px;
  }

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

  .field-error {
    color: #dc3545;
    font-size: 0.8rem;
    margin-top: 4px;
  }

  .msg-success {
    background: #d4edda;
    color: #155724;
    padding: 16px;
    border-radius: 8px;
    text-align: center;
    line-height: 1.8;
    white-space: pre-line;
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
</style>
