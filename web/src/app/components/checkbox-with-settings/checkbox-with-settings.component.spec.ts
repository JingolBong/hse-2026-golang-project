import {ComponentFixture, TestBed} from "@angular/core/testing";
import {HttpClientTestingModule} from "@angular/common/http/testing";

import {CheckboxWithSettingsComponent} from "./checkbox-with-settings.component";
import {ProjectServices} from "../../services/project.services";
import {CheckedSetting} from "../../models/check-setting.model";
import {SettingBox} from "../../models/setting.model";

describe("CheckboxWithSettingsComponent", () => {
  let component: CheckboxWithSettingsComponent;
  let fixture: ComponentFixture<CheckboxWithSettingsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [CheckboxWithSettingsComponent],
      imports: [HttpClientTestingModule],
      providers: [ProjectServices],
    })
      .overrideComponent(CheckboxWithSettingsComponent, {set: {template: ""}})
      .compileComponents();

    fixture = TestBed.createComponent(CheckboxWithSettingsComponent);
    component = fixture.componentInstance;
    component.box = new SettingBox("Some chart", true, 4);
  });

  it("should create and seed isChecked from the box state", () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
    expect(component.isChecked).toBeTrue();
  });

  it("changed() emits a CheckedSetting carrying the box text and id", () => {
    fixture.detectChanges();
    let emitted: CheckedSetting | undefined;
    component.onChecked.subscribe((e: CheckedSetting) => (emitted = e));

    component.isChecked = false;
    component.changed(false);

    expect(emitted).toEqual(new CheckedSetting("Some chart", false, 4));
  });
});
