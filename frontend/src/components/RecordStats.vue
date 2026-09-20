<script setup lang="ts">
import type { WorkRecord } from '../types/domain';

defineProps<{ records: WorkRecord[] }>();
</script>

<template>
  <section class="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
    <div class="mb-3 flex items-center justify-between">
      <h2 class="text-lg font-black">作业记录统计</h2>
      <el-button size="small">导出 Excel</el-button>
    </div>
    <el-table :data="records" size="small">
      <el-table-column prop="workDate" label="日期" width="105" />
      <el-table-column prop="machineCode" label="农机" width="120" />
      <el-table-column prop="driverName" label="驾驶员" width="85" />
      <el-table-column prop="taskType" label="类型" width="80" />
      <el-table-column prop="actualHours" label="工时" width="75" />
      <el-table-column prop="fuelLiters" label="油耗/L" width="85" />
      <el-table-column prop="areaMu" label="面积/亩" width="85" />
      <el-table-column prop="fuelCost" label="油耗成本" width="90" />
      <el-table-column label="关联任务" min-width="100">
        <template #default="{ row }">{{ row.taskId ?? '—' }}</template>
      </el-table-column>
    </el-table>
  </section>
</template>
