import { Options } from 'highcharts';

// Comparison version: one column series per project (legend on), no default series.
export const complexityTaskChartOptions: Options = {
  chart: {
    type: 'column',
  },
  credits: {
    enabled: false,
  },
  title: {
    text: 'Complexity task',
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
      text: 'Log time'
    }
  },

  plotOptions: {
    series: {
      borderRadius: 5,
    } as any,
  },

  series: [],
};
