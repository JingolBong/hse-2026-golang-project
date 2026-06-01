import { Options } from 'highcharts';

// Comparison version: one column series per project (legend on), no default series.
export const taskPriorityChartOptions: Options = {
  chart: {
    type: 'column',
  },
  credits: {
    enabled: false,
  },
  title: {
    text: 'Task priority',
  },
  yAxis: {
    visible: true,
    title: {
      text: 'Issue count'
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
