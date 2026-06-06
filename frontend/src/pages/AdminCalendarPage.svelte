<script>
  import { onMount } from 'svelte'
  import { api } from '../api/index.js'

  let events = []
  let loading = true
  let currentYear = new Date().getFullYear()
  let currentMonth = new Date().getMonth() + 1

  $: monthStr = `${currentYear}-${String(currentMonth).padStart(2, '0')}`

  onMount(async () => {
    await loadCalendar()
  })

  async function loadCalendar() {
    loading = true
    try {
      const res = await api.getCalendarEvents(monthStr)
      events = res.data || []
    } catch (e) {
      console.error(e)
    }
    loading = false
  }

  $: {
    monthStr
    loadCalendar()
  }

  function prevMonth() {
    currentMonth--
    if (currentMonth < 1) {
      currentMonth = 12
      currentYear--
    }
  }

  function nextMonth() {
    currentMonth++
    if (currentMonth > 12) {
      currentMonth = 1
      currentYear++
    }
  }

  function getDaysInMonth(year, month) {
    return new Date(year, month, 0).getDate()
  }

  function getFirstDayOfWeek(year, month) {
    return new Date(year, month - 1, 1).getDay()
  }

  $: daysInMonth = getDaysInMonth(currentYear, currentMonth)
  $: firstDay = getFirstDayOfWeek(currentYear, currentMonth)
  $: weekDays = ['日', '一', '二', '三', '四', '五', '六']

  function getEventsForDay(day) {
    const dateStr = `${currentYear}-${String(currentMonth).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    return events.filter(e => e.date === dateStr)
  }

  function isToday(day) {
    const today = new Date()
    return today.getFullYear() === currentYear && today.getMonth() + 1 === currentMonth && today.getDate() === day
  }
</script>

<div class="page">
  <div class="page-header">
    <h2>📅 排期日历</h2>
  </div>

  <div class="calendar-controls">
    <button class="btn btn-nav" on:click={prevMonth}>◀</button>
    <span class="month-label">{currentYear}年{currentMonth}月</span>
    <button class="btn btn-nav" on:click={nextMonth}>▶</button>
  </div>

  {#if loading}
    <div class="loading">加载中...</div>
  {:else}
    <div class="calendar">
      <div class="calendar-header">
        {#each weekDays as day}
          <div class="weekday">{day}</div>
        {/each}
      </div>
      <div class="calendar-body">
        {#each Array(firstDay) as _}
          <div class="day-cell empty"></div>
        {/each}
        {#each Array(daysInMonth) as _, i}
          {@const day = i + 1}
          {@const dayEvents = getEventsForDay(day)}
          <div class="day-cell" class:today={isToday(day)}>
            <div class="day-number">{day}</div>
            {#each dayEvents as event}
              <div class="event" class:full={event.is_full}>
                <span class="event-title">{event.title}</span>
                <span class="event-info">{event.time_slot}</span>
                <span class="event-capacity">{event.booked}/{event.capacity}</span>
              </div>
            {/each}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <div class="legend">
    <div class="legend-item">
      <span class="legend-dot open"></span> 报名中
    </div>
    <div class="legend-item">
      <span class="legend-dot full"></span> 已满员
    </div>
  </div>
</div>

<style>
  .page { max-width: 1100px; margin: 0 auto; }

  .page-header { margin-bottom: 20px; }
  .page-header h2 { font-size: 1.5rem; }

  .calendar-controls {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 20px;
    margin-bottom: 20px;
  }

  .btn {
    padding: 8px 16px;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.9rem;
  }

  .btn-nav {
    background: white;
    border: 1px solid #ddd;
    font-size: 1rem;
  }
  .btn-nav:hover { background: #f8f8f8; }

  .month-label {
    font-size: 1.2rem;
    font-weight: 600;
    min-width: 120px;
    text-align: center;
  }

  .loading {
    text-align: center;
    padding: 60px;
    color: #888;
  }

  .calendar {
    background: white;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
    overflow: hidden;
  }

  .calendar-header {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    background: #faf5f0;
  }

  .weekday {
    padding: 12px;
    text-align: center;
    font-size: 0.85rem;
    color: #8b7355;
    font-weight: 600;
  }

  .calendar-body {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
  }

  .day-cell {
    min-height: 100px;
    padding: 8px;
    border: 1px solid #f0e8e0;
    border-top: none;
    border-left: none;
  }

  .day-cell.empty { background: #fafafa; }

  .day-cell.today .day-number {
    background: #8b2500;
    color: white;
    border-radius: 50%;
    width: 26px;
    height: 26px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .day-number {
    font-size: 0.85rem;
    color: #666;
    margin-bottom: 4px;
  }

  .event {
    background: #d4edda;
    border-radius: 4px;
    padding: 3px 6px;
    margin-bottom: 3px;
    font-size: 0.72rem;
    line-height: 1.4;
    display: flex;
    flex-direction: column;
  }

  .event.full {
    background: #f8d7da;
  }

  .event-title { font-weight: 600; }
  .event-info { color: #666; }
  .event-capacity { color: #888; }

  .legend {
    display: flex;
    gap: 20px;
    margin-top: 16px;
    justify-content: center;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.85rem;
    color: #666;
  }

  .legend-dot {
    width: 12px;
    height: 12px;
    border-radius: 3px;
  }

  .legend-dot.open { background: #d4edda; }
  .legend-dot.full { background: #f8d7da; }
</style>
