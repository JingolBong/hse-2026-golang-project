import { Options } from 'highcharts';

// Comparison version: one column series per project (legend on), no default series.
export const closeTaskPriorityChartOptions: Options = {
  chart: {
    type: 'column',
  },
  credits: {
    enabled: false,
  },
  title: {
    text: 'Close task priority',
  },
  yAxis: {
    visible: true,
    title: {
      text: 'Close issue count'
    }
  },
  legend: {
    enabled: true,
  },
  xAxis: {
    lineColor: '#fff',
    categories: [],
    title: {
      text: 'Priority'
    }
  },

  plotOptions: {
    series: {
      borderRadius: 5,
    } as any,
  },

  series: [],
};
