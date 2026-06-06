<script>
  import { onMount } from 'svelte'
  import { api } from '../api/index.js'

  let stats = []
  let loading = true

  onMount(async () => {
    await loadStats()
  })

  async function loadStats() {
    loading = true
    try {
      const res = await api.getStats()
      stats = res.data || []
    } catch (e) {
      console.error(e)
    }
    loading = false
  }

  let totalCourses = 0
  let totalBooked = 0
  let totalAttended = 0
  let avgRate = 0

  $: {
    totalCourses = stats.length
    totalBooked = stats.reduce((s, c) => s + c.booked, 0)
    totalAttended = stats.reduce((s, c) => s + c.attended, 0)
    avgRate = totalBooked > 0 ? (totalAttended / totalBooked * 100).toFixed(1) : 0
  }
</script>

<div class="page">
  <div class="page-header">
    <h2>📊 预约统计</h2>
    <button class="btn btn-primary" on:click={loadStats}>🔄 刷新</button>
  </div>

  {#if loading}
    <div class="loading">加载中...</div>
  {:else}
    <div class="summary-cards">
      <div class="summary-card">
        <div class="summary-value">{totalCourses}</div>
        <div class="summary-label">课程总数</div>
      </div>
      <div class="summary-card">
        <div class="summary-value">{totalBooked}</div>
        <div class="summary-label">总预约人次</div>
      </div>
      <div class="summary-card">
        <div class="summary-value">{totalAttended}</div>
        <div class="summary-label">总到课人次</div>
      </div>
      <div class="summary-card highlight">
        <div class="summary-value">{avgRate}%</div>
        <div class="summary-label">平均到课率</div>
      </div>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>课程名称</th>
            <th>日期</th>
            <th>时段</th>
            <th>名额</th>
            <th>已预约</th>
            <th>已到课</th>
            <th>到课率</th>
          </tr>
        </thead>
        <tbody>
          {#each stats as s}
            <tr>
              <td><strong>{s.course_title}</strong></td>
              <td>{s.course_date}</td>
              <td>{s.course_slot}</td>
              <td>{s.capacity}</td>
              <td>{s.booked}</td>
              <td>{s.attended}</td>
              <td>
                <div class="rate-cell">
                  <div class="rate-bar">
                    <div class="rate-fill" style="width: {s.booked > 0 ? (s.attended / s.booked * 100) : 0}%"></div>
                  </div>
                  <span class="rate-text">{s.attend_rate}%</span>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

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
    padding: 8px 16px;
    background: white;
    border: 1px solid #ddd;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.85rem;
  }

  .btn:hover { background: #f8f8f8; }
  .btn-primary { background: #8b2500; color: white; border-color: #8b2500; }
  .btn-primary:hover { background: #a0522d; }

  .loading {
    text-align: center;
    padding: 60px;
    color: #888;
  }

  .summary-cards {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-bottom: 24px;
  }

  .summary-card {
    background: white;
    border-radius: 12px;
    padding: 24px;
    text-align: center;
    box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  }

  .summary-card.highlight {
    background: linear-gradient(135deg, #8b2500, #a0522d);
    color: white;
  }

  .summary-value {
    font-size: 2rem;
    font-weight: 700;
    margin-bottom: 4px;
  }

  .summary-label {
    font-size: 0.85rem;
    opacity: 0.7;
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

  .rate-cell {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .rate-bar {
    flex: 1;
    height: 8px;
    background: #f0e8e0;
    border-radius: 4px;
    overflow: hidden;
  }

  .rate-fill {
    height: 100%;
    background: #28a745;
    border-radius: 4px;
    transition: width 0.3s;
  }

  .rate-text {
    font-size: 0.85rem;
    font-weight: 600;
    min-width: 48px;
    color: #28a745;
  }
</style>
