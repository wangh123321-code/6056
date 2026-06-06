<script>
  import { onMount } from 'svelte'
  import { api } from '../api/index.js'

  export let userPhone = ''

  let phone = ''
  let bookings = []
  let loading = false
  let searched = false
  let cancelLoading = {}

  $: if (userPhone && !searched) {
    phone = userPhone
    searchBookings()
  }

  async function searchBookings() {
    if (!phone.trim()) return
    loading = true
    searched = true
    try {
      const res = await api.getMyBookings(phone.trim())
      bookings = res.data || []
    } catch (e) {
      console.error(e)
      bookings = []
    }
    loading = false
  }

  async function cancelBooking(booking) {
    if (!confirm(`确认取消「${booking.course_title}」的预约吗？`)) return
    cancelLoading[booking.id] = true
    try {
      await api.cancelBooking(booking.id)
      await searchBookings()
    } catch (e) {
      alert(e.message)
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
    <input
      type="tel"
      bind:value={phone}
      placeholder="请输入手机号查询预约记录"
      on:keydown={(e) => e.key === 'Enter' && searchBookings()}
    />
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
        <div class="booking-card" class:cancelled={booking.status === 'cancelled'}>
          <div class="booking-status">
            {#if booking.status === 'booked'}
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
              <span>预约时间：{formatDate(booking.created_at)}</span>
              {#if booking.cancelled_at}
                <span>取消时间：{formatDate(booking.cancelled_at)}</span>
              {/if}
            </div>
          </div>
          {#if booking.status === 'booked'}
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
  }

  .search-bar input {
    flex: 1;
    padding: 10px 14px;
    border: 1px solid #ddd;
    border-radius: 8px;
    font-size: 0.9rem;
  }

  .search-bar input:focus {
    outline: none;
    border-color: #8b2500;
  }

  .btn {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-size: 0.9rem;
    cursor: pointer;
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
  }

  .booking-card.cancelled { opacity: 0.65; }

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
