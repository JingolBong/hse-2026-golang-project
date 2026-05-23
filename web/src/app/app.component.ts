import {Component} from "@angular/core";
import {ConfigurationService} from "./services/configuration.services";

@Component({
  selector: "app-root",
  templateUrl: "./app.component.html",
  styleUrls: ["./app.component.css"],
})
export class AppComponent {
  title = "web";
  webUrl = "";

  constructor(configurationService: ConfigurationService) {
    const host = configurationService.getValue<string>("webHost", "localhost");
    const port = configurationService.getValue<number>("webPort", 4200);
    this.webUrl = `${host}:${port}`;
  }
}
