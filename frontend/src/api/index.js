const BASE = '/api'

async function request(url, options = {}) {
  const res = await fetch(BASE + url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || '请求失败')
  return data
}

export const api = {
  getCourses(params = '') {
    return request('/courses' + (params ? '?' + params : ''))
  },
  getCourse(id) {
    return request('/courses/' + id)
  },
  createCourse(data) {
    return request('/courses', { method: 'POST', body: JSON.stringify(data) })
  },
  updateCourse(id, data) {
    return request('/courses/' + id, { method: 'PUT', body: JSON.stringify(data) })
  },
  deleteCourse(id) {
    return request('/courses/' + id, { method: 'DELETE' })
  },
  createBooking(data) {
    return request('/bookings', { method: 'POST', body: JSON.stringify(data) })
  },
  getMyBookings(phone) {
    return request('/bookings/my?phone=' + encodeURIComponent(phone))
  },
  cancelBooking(id) {
    return request('/bookings/' + id, { method: 'DELETE' })
  },
  getCourseBookings(courseId) {
    return request('/courses/' + courseId + '/bookings')
  },
  getCourseWaitlist(courseId) {
    return request('/courses/' + courseId + '/waitlist')
  },
  checkAvailability(courseId) {
    return request('/courses/' + courseId + '/availability')
  },
  getCalendarEvents(month) {
    return request('/courses/calendar?month=' + month)
  },
  getStats() {
    return request('/stats')
  },
  markAttendance(bookingId) {
    return request('/attendance/' + bookingId, { method: 'PUT' })
  },
  addToWaitlist(data) {
    return request('/waitlist', { method: 'POST', body: JSON.stringify(data) })
  },
  confirmWaitlist(waitlistId) {
    return request('/waitlist/confirm', { method: 'POST', body: JSON.stringify({ waitlist_id: waitlistId }) })
  },
  cancelWaitlist(id) {
    return request('/waitlist/' + id, { method: 'DELETE' })
  },
}
