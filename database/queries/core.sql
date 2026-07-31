-- name: ListStations :many
SELECT id, code, name FROM stations WHERE active ORDER BY name;

-- name: ListRouteStations :many
SELECT s.id, s.code, s.name, rs.position, rs.cumulative_distance_m
FROM route_stations rs JOIN stations s ON s.id = rs.station_id
WHERE rs.route_id = $1 AND s.active ORDER BY rs.position;

-- name: ListTrainRuns :many
SELECT id, train_id, route_id, service_date, departure_at, arrival_at, status
FROM train_runs WHERE service_date = $1 AND status = 'SCHEDULED' ORDER BY departure_at;
