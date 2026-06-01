import { Component, OnInit } from '@angular/core';
import {ActivatedRoute} from '@angular/router'
import {DatabaseProjectServices} from "../services/database-project.services";
import {Chart} from "angular-highcharts"
import {Options} from "highcharts";
import {openTaskChartOptions} from "./helpers/openTaskChartOptions";
import {complexityTaskChartOptions} from "./helpers/complexityTaskChartOptions";
import {taskPriorityChartOptions} from "./helpers/taskPriorityChartOptions";
import {closeTaskPriorityChartOptions} from "./helpers/closeTaskPriorityChartOptions";
import {ConfigurationService} from "../services/configuration.services";


@Component({
  selector: 'app-compare-project-page',
  templateUrl: './compare-project-page.component.html',
  styleUrls: ['./compare-project-page.component.scss']
})

export class CompareProjectPageComponent implements OnInit {
  projects: string[] = []
  ids: string[] = []
  resultReq: ReqData[] = []
  openTaskChart = new Chart()
  complexityTaskChart = new Chart()
  taskPriorityChart = new Chart()
  closeTaskPriorityChart = new Chart()

  colors = ["blue", "green", "red", "orange", "purple", "black"]

  webUrl = ""

  constructor(configurationService: ConfigurationService,
              private route: ActivatedRoute,
              private dbProjectService: DatabaseProjectServices) {
    this.projects = this.route.snapshot.queryParamMap.getAll("keys")
    this.ids = this.route.snapshot.queryParamMap.getAll("value")
    const host = configurationService.getValue<string>("webHost", "localhost");
    const port = configurationService.getValue<number>("webPort", 4200);
    this.webUrl = `${host}:${port}`;
  }



  ngOnInit(): void {
    for (let i = 0; i < this.projects.length; i++) {
      this.dbProjectService.getProjectStatByID(this.ids[i]).subscribe(projects => {
        this.resultReq[i] = projects.data
      })
    }

    let openTaskElem = document.getElementById('open-task') as HTMLElement;
    let openTaskTitle = document.getElementById('open-task-title') as HTMLElement;
    this.dbProjectService.getComplitedGraph("1", this.projects).subscribe(info => {
      if (info.data["count"] == null) {
        openTaskElem.remove()
        openTaskTitle.remove()
      }
      else{
        // @ts-ignore
        openTaskChartOptions.xAxis["categories"] = info.data["categories"]
        for (let j = 0; j < this.projects.length; j++){
          var count = []
          for (let i = 0; i < info.data["categories"].length; i++){
            // @ts-ignore
            count.push(info.data["count"][info.data["categories"][i]][j])
          }
          openTaskChartOptions.series?.push({ name: this.projects[j],
            type: "column",
            color: this.colors[j],
            data: count})
          this.openTaskChart = new Chart(openTaskChartOptions)
        }
      }
    })

    // Histogram-shaped comparisons (fixed categories): one column series per
    // project, pulled from each project's single-project graph. No backend
    // compare endpoint is needed for these.
    this.buildComparison("4", complexityTaskChartOptions, c => (this.complexityTaskChart = c))
    this.buildComparison("5", taskPriorityChartOptions, c => (this.taskPriorityChart = c))
    this.buildComparison("6", closeTaskPriorityChartOptions, c => (this.closeTaskPriorityChart = c))
  }

  // Fetch each project's single-project graph for `task` and overlay one column
  // series per project onto the shared options, rebuilding the chart as
  // responses arrive. Projects with no data for the task are skipped.
  private buildComparison(task: string, options: Options, assign: (chart: Chart) => void): void {
    for (let j = 0; j < this.projects.length; j++) {
      const projectName = this.projects[j]
      this.dbProjectService.getGraph(task, projectName).subscribe(info => {
        const data: any = info.data
        if (data == null || data["categories"] == null) {
          return
        }
        // @ts-ignore
        options.xAxis["categories"] = data["categories"]
        const count: number[] = []
        for (let i = 0; i < data["categories"].length; i++) {
          count.push(data["count"][data["categories"][i]])
        }
        options.series?.push({
          name: projectName,
          type: "column",
          color: this.colors[j],
          data: count,
        } as any)
        assign(new Chart(options))
      })
    }
  }

  ngOnDestroy(): void{
    // @ts-ignore
    openTaskChartOptions.xAxis["categories"] = []
    openTaskChartOptions.series = []
    // @ts-ignore
    complexityTaskChartOptions.xAxis["categories"] = []
    complexityTaskChartOptions.series = []
    // @ts-ignore
    taskPriorityChartOptions.xAxis["categories"] = []
    taskPriorityChartOptions.series = []
    // @ts-ignore
    closeTaskPriorityChartOptions.xAxis["categories"] = []
    closeTaskPriorityChartOptions.series = []
  }
}



class ReqData {
  Id: number;
  Key: string;
  Name: string;
  allIssuesCount: number;
  averageIssuesCount: string;
  averageTime: number;
  closeIssuesCount: number;
  openIssuesCount: number;
  resolvedIssuesCount: number;
  reopenedIssuesCount: number;
  progressIssuesCount: number;
}
