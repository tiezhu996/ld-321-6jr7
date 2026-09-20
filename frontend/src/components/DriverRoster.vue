<script setup lang="ts">
import { STATUS_COLORS } from '../constants/app.constants';
import type { Driver } from '../types/domain';

defineProps<{ drivers: Driver[] }>();
</script>

<template>
  <section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
    <h2 class="mb-3 text-lg font-black">驾驶员管理</h2>
    <div class="grid gap-3 md:grid-cols-3">
      <article v-for="driver in drivers" :key="driver.id" class="rounded-md border border-slate-200 p-3">
        <div class="flex items-center justify-between">
          <strong>{{ driver.name }}</strong>
          <el-tag :type="STATUS_COLORS[driver.status] ?? 'info'" size="small">{{ driver.status }}</el-tag>
        </div>
        <p class="mt-2 text-sm text-slate-600">{{ driver.licenseNo }}</p>
        <p class="text-sm text-slate-600">{{ driver.phone }}</p>
        <div class="mt-2 flex items-center justify-between text-sm text-emerald-700">
          <span>{{ driver.shift }} · 休 {{ driver.restDay }} · {{ driver.monthAreaMu }} 亩</span>
          <el-rate :model-value="driver.rating" disabled size="small" />
        </div>
      </article>
    </div>
  </section>
</template>
