<script setup lang="ts">
import { computed, ref } from 'vue';
import { STATUS_COLORS } from '../constants/app.constants';
import type { Machine, MaintenanceReminder } from '../types/domain';

const props = defineProps<{ machines: Machine[]; maintenance?: MaintenanceReminder[] }>();

const statusFilter = ref('all');

// 取该农机所有保养提醒中的最小剩余工时，作为“保养剩余时长”。
const remainingHoursOf = (code: string): number | null => {
  const list = (props.maintenance ?? []).filter((item) => item.machineCode === code);
  if (list.length === 0) {
    return null;
  }
  return Math.min(...list.map((item) => item.remainingHours));
};

const filteredMachines = computed(() => {
  if (statusFilter.value === 'all') {
    return props.machines;
  }
  const mapping: Record<string, string> = { idle: '空闲', working: '作业中', repair: '维修中' };
  return props.machines.filter((m) => m.status === mapping[statusFilter.value]);
});
</script>

<template>
  <section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
    <div class="mb-3 flex items-center justify-between">
      <h2 class="text-lg font-black">农机档案</h2>
      <el-select v-model="statusFilter" placeholder="按状态筛选" size="small" class="w-36">
        <el-option label="全部" value="all" />
        <el-option label="空闲" value="idle" />
        <el-option label="作业中" value="working" />
        <el-option label="维修中" value="repair" />
      </el-select>
    </div>
    <el-table :data="filteredMachines" size="small">
      <el-table-column prop="code" label="编号" width="120" />
      <el-table-column prop="name" label="农机" />
      <el-table-column prop="model" label="型号" width="110" />
      <el-table-column prop="horsepower" label="马力" width="80" />
      <el-table-column prop="field" label="所属地块" />
      <el-table-column label="累计工时" width="90">
        <template #default="{ row }">{{ row.workHours }}h</template>
      </el-table-column>
      <el-table-column label="保养剩余" width="100">
        <template #default="{ row }">
          <span :class="remainingHoursOf(row.code) === 0 ? 'font-bold text-red-600' : 'text-slate-700'">
            {{ remainingHoursOf(row.code) === null ? '—' : `${remainingHoursOf(row.code)}h` }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="STATUS_COLORS[row.status]" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="qrCode" label="二维码" width="110" />
    </el-table>
  </section>
</template>
