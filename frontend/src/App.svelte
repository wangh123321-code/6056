<script>
  import { onMount, setContext } from 'svelte'
  import BookingPage from './pages/BookingPage.svelte'
  import MyBookingsPage from './pages/MyBookingsPage.svelte'
  import AdminCoursesPage from './pages/AdminCoursesPage.svelte'
  import AdminStatsPage from './pages/AdminStatsPage.svelte'
  import AdminCalendarPage from './pages/AdminCalendarPage.svelte'
  import Modal from './components/Modal.svelte'

  let currentRoute = 'booking'
  let userPhone = ''
  let userName = ''

  let modalShow = false
  let modalTitle = ''
  let modalMessage = ''
  let modalType = 'alert'
  let modalConfirmText = '确定'
  let modalCancelText = '取消'
  let modalResolve = null

  const routes = [
    { id: 'booking', label: '课程预约', icon: '📋' },
    { id: 'my-bookings', label: '我的预约', icon: '📋' },
    { id: 'admin-courses', label: '课程管理', icon: '⚙️' },
    { id: 'admin-calendar', label: '排期日历', icon: '📅' },
    { id: 'admin-stats', label: '预约统计', icon: '📊' },
  ]

  onMount(() => {
    const savedPhone = localStorage.getItem('workshop_user_phone')
    const savedName = localStorage.getItem('workshop_user_name')
    if (savedPhone) userPhone = savedPhone
    if (savedName) userName = savedName
  })

  $: if (userPhone) localStorage.setItem('workshop_user_phone', userPhone)
  $: if (userName) localStorage.setItem('workshop_user_name', userName)

  function showAlert(title, message) {
    return new Promise((resolve) => {
      modalTitle = title
      modalMessage = message
      modalType = 'alert'
      modalConfirmText = '确定'
      modalShow = true
      modalResolve = resolve
    })
  }

  function showConfirm(title, message, confirmText = '确定', cancelText = '取消') {
    return new Promise((resolve) => {
      modalTitle = title
      modalMessage = message
      modalType = 'confirm'
      modalConfirmText = confirmText
      modalCancelText = cancelText
      modalShow = true
      modalResolve = resolve
    })
  }

  setContext('dialog', { showAlert, showConfirm })

  function navigate(route) {
    currentRoute = route
  }

  function handleModalResolve(val) {
    if (modalResolve) modalResolve(val)
    modalResolve = null
  }
</script>

<main>
  <nav class="sidebar">
    <div class="logo">
      <span class="logo-icon">✂️</span>
      <h1>剪纸工坊</h1>
      <p class="subtitle">课程预约系统</p>
    </div>

    <div class="nav-section">
      <p class="nav-label">用户端</p>
      {#each routes.slice(0, 2) as route}
        <button class="nav-btn" class:active={currentRoute === route.id} on:click={() => navigate(route.id)}>
          <span class="nav-icon">{route.icon}</span>
          {route.label}
        </button>
      {/each}
    </div>

    <div class="nav-section">
      <p class="nav-label">管理端</p>
      {#each routes.slice(2) as route}
        <button class="nav-btn" class:active={currentRoute === route.id} on:click={() => navigate(route.id)}>
          <span class="nav-icon">{route.icon}</span>
          {route.label}
        </button>
      {/each}
    </div>
  </nav>

  <div class="content">
    {#if currentRoute === 'booking'}
      <BookingPage bind:userPhone bind:userName />
    {:else if currentRoute === 'my-bookings'}
      <MyBookingsPage bind:userPhone />
    {:else if currentRoute === 'admin-courses'}
      <AdminCoursesPage />
    {:else if currentRoute === 'admin-calendar'}
      <AdminCalendarPage />
    {:else if currentRoute === 'admin-stats'}
      <AdminStatsPage />
    {/if}
  </div>
</main>

<Modal
  bind:show={modalShow}
  title={modalTitle}
  message={modalMessage}
  type={modalType}
  confirmText={modalConfirmText}
  cancelText={modalCancelText}
  resolve={handleModalResolve}
/>

<style>
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
    background: #f5f0eb;
    color: #2c1810;
    min-height: 100vh;
  }

  main {
    display: flex;
    min-height: 100vh;
  }

  .sidebar {
    width: 240px;
    background: linear-gradient(180deg, #8b2500 0%, #a0522d 100%);
    color: white;
    padding: 24px 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    box-shadow: 2px 0 12px rgba(0,0,0,0.15);
  }

  .logo {
    text-align: center;
    padding-bottom: 20px;
    border-bottom: 1px solid rgba(255,255,255,0.2);
    margin-bottom: 8px;
  }

  .logo-icon {
    font-size: 2.5rem;
    display: block;
    margin-bottom: 8px;
  }

  .logo h1 {
    font-size: 1.4rem;
    font-weight: 700;
    letter-spacing: 2px;
  }

  .subtitle {
    font-size: 0.8rem;
    opacity: 0.8;
    margin-top: 4px;
  }

  .nav-section {
    margin-top: 8px;
  }

  .nav-label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    opacity: 0.6;
    padding: 4px 12px;
    margin-top: 8px;
  }

  .nav-btn {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 10px 12px;
    background: transparent;
    border: none;
    color: rgba(255,255,255,0.85);
    font-size: 0.9rem;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.2s;
  }

  .nav-btn:hover {
    background: rgba(255,255,255,0.15);
    color: white;
  }

  .nav-btn.active {
    background: rgba(255,255,255,0.25);
    color: white;
    font-weight: 600;
  }

  .nav-icon {
    font-size: 1.1rem;
  }

  .content {
    flex: 1;
    padding: 32px;
    overflow-y: auto;
  }
</style>
