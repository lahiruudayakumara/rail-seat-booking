export interface ScheduleStationData {
  position: number;
  code: string;
  name: string;
  dist: string;
}

export const STATIONS_SCHEDULE_DATA: ScheduleStationData[] = [
  { position: 0, code: "FOT", name: "Colombo Fort", dist: "0 km" },
  { position: 1, code: "RGM", name: "Ragama", dist: "14 km" },
  { position: 2, code: "GPH", name: "Gampaha", dist: "28 km" },
  { position: 3, code: "PDA", name: "Peradeniya Junction", dist: "115 km" },
  { position: 4, code: "KDY", name: "Kandy", dist: "120 km" },
  { position: 5, code: "NAN", name: "Nanu Oya", dist: "207 km" },
  { position: 6, code: "ELL", name: "Ella", dist: "271 km" },
  { position: 7, code: "BAD", name: "Badulla", dist: "292 km" },
];
