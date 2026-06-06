<script>
  import { getContext } from 'svelte'
  import { api } from '../api/index.js'

  export let userPhone = ''

  const { showAlert, showConfirm } = getContext('dialog')

  let phone = ''
  let bookings = []
  let loading = false
  let searched = false
  let cancelLoading = {}
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

  $: if (phone) validatePhone(phone)

  async function searchBookings() {
    if (!validatePhone(phone.trim())) {
      await showAlert('输入有误', phoneError)
      return
    }
    loading = true
    searched = true
    try {
      const res = await api.getMyBookings(phone.trim())
      bookings = res.data || []
      userPhone = phone.trim()
    } catch (e) {
      await showAlert('查询失败', e.message)
      bookings = []
    }
    loading = false
  }

  async function cancelBooking(booking) {
    const confirmed = await showConfirm(
      '确认取消',
      `确认取消「${booking.course_title}」\n（${booking.course_date} ${booking.course_slot}）的预约吗？\n取消后名额将释放给其他用户。`,
      '确认取消',
      '再想想'
    )
    if (!confirmed) return

    cancelLoading[booking.id] = true
    try {
      await api.cancelBooking(booking.id)
      await showAlert('取消成功', '预约已取消，名额已释放')
      await searchBookings()
    } catch (e) {
      await showAlert('取消失败', e.message)
    }
    cancelLoading[booking.id] = false
  }

  async function cancelWaitlist(booking) {
    const confirmed = await showConfirm(
      '确认退出',
      `确认退出「${booking.course_title}」\n（${booking.course_date} ${booking.course_slot}）的候补队列吗？`,
      '确认退出',
      '再想想'
    )
    if (!confirmed) return

    cancelLoading[booking.id] = true
    try {
      await api.cancelWaitlist(booking.id)
      await showAlert('已退出', '已退出候补队列')
      await searchBookings()
    } catch (e) {
      await showAlert('操作失败', e.message)
    }
    cancelLoading[booking.id] = false
  }

  async function confirmWaitlist(booking) {
    const confirmed = await showConfirm(
      '确认候补',
      `有名额释放！确认预约「${booking.course_title}」\n（${booking.course_date} ${booking.course_slot}）吗？\n请在15分钟内确认，超时将自动顺延给下一位。`,
      '确认预约',
      '暂不确认'
    )
    if (!confirmed) return

    cancelLoading[booking.id] = true
    try {
      const res = await api.confirmWaitlist(booking.id)
      await showAlert('预约成功', `已成功预约「${res.data.course_title}」\n${res.data.date} ${res.data.time_slot}`)
      await searchBookings()
    } catch (e) {
      await showAlert('确认失败', e.message)
    }
    cancelLoading[booking.id] = false
  }

  function formatDate(dt) {
    if (!dt) return '-'
    return new Date(dt).toLocaleString('zh-CN')
  }
</script>

<div class="page">
  <div class="page-header">
    <h2>📋 我的预约</h2>
  </div>

  <div class="search-bar">
    <div class="search-input-wrap">
      <input
        type="tel"
        bind:value={phone}
        placeholder="请输入11位手机号查询预约记录"
        maxlength="11"
        on:keydown={(e) => e.key === 'Enter' && searchBookings()}
      />
      {#if phoneError}
        <div class="field-error">{phoneError}</div>
      {/if}
    </div>
    <button class="btn btn-primary" on:click={searchBookings} disabled={loading}>
      {loading ? '查询中...' : '查询'}
    </button>
  </div>

  {#if loading}
    <div class="loading">加载中...</div>
  {:else if searched && bookings.length === 0}
    <div class="empty">暂无预约记录</div>
  {:else if bookings.length > 0}
    <div class="booking-list">
      {#each bookings as booking}
        <div class="booking-card" class:cancelled={booking.status === 'cancelled'} class:waitlist={booking.type === 'waitlist'}>
          <div class="booking-status">
            {#if booking.type === 'waitlist'}
              {#if booking.status === 'notified'}
                <span class="status-badge status-notified">待确认</span>
              {:else}
                <span class="status-badge status-waiting">候补中 · 第{booking.position}位</span>
              {/if}
            {:else if booking.status === 'booked'}
              <span class="status-badge status-booked">已预约</span>
            {:else if booking.status === 'attended'}
              <span class="status-badge status-attended">已到课</span>
            {:else}
              <span class="status-badge status-cancelled">已取消</span>
            {/if}
          </div>
          <div class="booking-info">
            <h3>{booking.course_title}</h3>
            <div class="detail-row">
              <span>📅 {booking.course_date}</span>
              <span>🕐 {booking.course_slot}</span>
            </div>
            <div class="detail-row">
              <span>👤 {booking.user_name}</span>
              <span>📱 {booking.user_phone}</span>
            </div>
            <div class="detail-row time">
              <span>{booking.type === 'waitlist' ? '候补时间' : '预约时间'}：{formatDate(booking.created_at)}</span>
              {#if booking.cancelled_at}
                <span>取消时间：{formatDate(booking.cancelled_at)}</span>
              {/if}
              {#if booking.type === 'waitlist' && booking.status === 'notified' && booking.expires_at}
                <span class="expire-warning">⏰ 请在 {formatDate(booking.expires_at)} 前确认</span>
              {/if}
            </div>
          </div>
          {#if booking.type === 'waitlist'}
            <div class="booking-action">
              {#if booking.status === 'notified'}
                <div class="action-buttons">
                  <button
                    class="btn btn-primary"
                    on:click={() => confirmWaitlist(booking)}
                    disabled={cancelLoading[booking.id]}
                  >
                    {cancelLoading[booking.id] ? '处理中...' : '确认预约'}
                  </button>
                  <button
                    class="btn btn-danger"
                    on:click={() => cancelWaitlist(booking)}
                    disabled={cancelLoading[booking.id]}
                  >
                    退出候补
                  </button>
                </div>
              {:else if booking.status === 'waiting'}
                <button
                  class="btn btn-danger"
                  on:click={() => cancelWaitlist(booking)}
                  disabled={cancelLoading[booking.id]}
                >
                  {cancelLoading[booking.id] ? '取消中...' : '退出候补'}
                </button>
              {/if}
            </div>
          {:else if booking.status === 'booked'}
            <div class="booking-action">
              <button
                class="btn btn-danger"
                on:click={() => cancelBooking(booking)}
                disabled={cancelLoading[booking.id]}
              >
                {cancelLoading[booking.id] ? '取消中...' : '取消预约'}
              </button>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    max-width: 800px;
    margin: 0 auto;
  }

  .page-header {
    margin-bottom: 24px;
  }

  .page-header h2 { font-size: 1.5rem; }

  .search-bar {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
    align-items: flex-start;
  }

  .search-input-wrap {
    flex: 1;
  }

  .search-bar input {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid #ddd;
    border-radius: 8px;
    font-size: 0.9rem;
  }

  .search-bar input:focus {
    outline: none;
    border-color: #8b2500;
  }

  .field-error {
    color: #dc3545;
    font-size: 0.8rem;
    margin-top: 4px;
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    cursor: pointer;
    transition: all 0.2s;
    height: fit-content;
  }

  .btn-primary { background: #8b2500; color: white; }
  .btn-primary:hover { background: #a0522d; }
  .btn-primary:disabled { background: #ccc; }

  .btn-danger {
    background: white;
    color: #dc3545;
    border: 1px solid #dc3545;
  }
  .btn-danger:hover { background: #dc3545; color: white; }
  .btn-danger:disabled { opacity: 0.5; }

  .loading, .empty {
    text-align: center;
    padding: 60px;
    color: #888;
    font-size: 1.1rem;
  }

  .booking-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .booking-card {
    background: white;
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
    display: flex;
    align-items: center;
    gap: 20px;
    animation: fadeIn 0.3s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .booking-card.cancelled { opacity: 0.65; }
  .booking-card.waitlist { border-left: 4px solid #fd7e14; }

  .status-badge {
    display: inline-block;
    padding: 4px 12px;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;
    white-space: nowrap;
  }

  .status-booked { background: #d4edda; color: #155724; }
  .status-attended { background: #cce5ff; color: #004085; }
  .status-cancelled { background: #f8d7da; color: #721c24; }
  .status-waiting { background: #ffe5cc; color: #a04000; }
  .status-notified { background: #ffc107; color: #523e02; animation: pulse 2s infinite; }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
  }

  .action-buttons {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .expire-warning {
    color: #dc3545;
    font-weight: 600;
  }

  .booking-info { flex: 1; }
  .booking-info h3 { font-size: 1.05rem; margin-bottom: 8px; }

  .detail-row {
    display: flex;
    gap: 20px;
    font-size: 0.85rem;
    color: #666;
    margin-top: 4px;
  }

  .detail-row.time {
    color: #aaa;
    font-size: 0.8rem;
    margin-top: 8px;
  }

  .booking-action {
    flex-shrink: 0;
  }
</style>
