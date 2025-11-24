<template>
  <div class="container mt-4">
    <h1>Dashboard</h1>
    <div class="row mt-4">
      <div class="col-md-4">
        <div class="card text-white bg-primary">
          <div class="card-body">
            <h5 class="card-title">Total Customers</h5>
            <h2>{{ stats.total_customers || 0 }}</h2>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card text-white bg-success">
          <div class="card-body">
            <h5 class="card-title">Total Accounts</h5>
            <h2>{{ stats.total_accounts || 0 }}</h2>
          </div>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card text-white bg-info">
          <div class="card-body">
            <h5 class="card-title">Active Accounts</h5>
            <h2>{{ stats.active_accounts || 0 }}</h2>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import apiClient from '../api/client'

export default {
  name: 'Dashboard',
  setup() {
    const stats = ref({})

    const loadStats = async () => {
      try {
        const response = await apiClient.get('/analytics')
        stats.value = response.data
      } catch (error) {
        console.error('Failed to load stats:', error)
      }
    }

    onMounted(() => {
      loadStats()
    })

    return {
      stats
    }
  }
}
</script>

