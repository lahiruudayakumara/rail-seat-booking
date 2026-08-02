import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import axios from "axios";
import { Link } from "react-router-dom";
import { getAdminDashboard, getRoutes, getTrainRuns } from "@/api";
import { Button, Clock, RefreshCw, ShieldCheck, Ticket, TrainFront, Users } from "@/components";
import type { AdminDashboard, Route, TrainRun } from "@/types";

const SESSION_KEY = "rail-admin-session-key";

function colomboDate() {
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Colombo",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date());
}

function money(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
  }).format(amountMinor / 100);
}

function MetricCard({ label, value, detail, icon }: { label: string; value: string; detail: string; icon: ReactNode }) {
  return (
    <article className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs font-bold uppercase tracking-wider text-stone-500">{label}</p>
          <p className="mt-2 font-heading text-3xl font-extrabold text-stone-900">{value}</p>
          <p className="mt-1 text-xs text-stone-500">{detail}</p>
        </div>
        <span className="rounded-xl bg-red-50 p-3 text-[#6b1724]">{icon}</span>
      </div>
    </article>
  );
}

function Login({ onLogin }: { onLogin: (key: string) => void }) {
  const [key, setKey] = useState("");
  const submit = (event: FormEvent) => {
    event.preventDefault();
    const value = key.trim();
    if (value) onLogin(value);
  };
  return (
    <main className="flex min-h-screen items-center justify-center bg-stone-100 px-5 py-12">
      <div className="w-full max-w-md rounded-3xl border border-stone-200 bg-white p-8 shadow-xl shadow-stone-300/30">
        <span className="inline-flex rounded-2xl bg-[#540d17] p-4 text-white"><ShieldCheck size={28} /></span>
        <p className="mt-6 text-xs font-bold uppercase tracking-[0.2em] text-[#851e2e]">Department access</p>
        <h1 className="mt-2 font-heading text-3xl font-extrabold">Operations dashboard</h1>
        <p className="mt-3 text-sm leading-6 text-stone-600">Use the administrator credential supplied through the secure operations channel. It is retained only for this browser session.</p>
        <form className="mt-7" onSubmit={submit}>
          <label className="field">
            <span>Administrator key</span>
            <input type="password" autoComplete="current-password" value={key} onChange={(event) => setKey(event.target.value)} required aria-label="Administrator key" />
          </label>
          <Button className="mt-4 w-full" type="submit">Open dashboard</Button>
        </form>
        <Link className="mt-5 block text-center text-sm font-semibold text-stone-500 hover:text-[#6b1724]" to="/">Return to passenger booking</Link>
      </div>
    </main>
  );
}

function DashboardContent({ dashboard }: { dashboard: AdminDashboard }) {
  const capacity = Math.max(dashboard.sellableSeatSegments, 1);
  const bookingTotal = Math.max(dashboard.confirmedBookings + dashboard.cancelledBookings + dashboard.heldBookings, 1);
  return (
    <>
      <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="Segment utilization" value={`${dashboard.segmentUtilizationPercent.toFixed(2)}%`} detail={`${dashboard.occupiedSeatSegments.toLocaleString()} of ${dashboard.sellableSeatSegments.toLocaleString()} seat-segments`} icon={<TrainFront size={22} />} />
        <MetricCard label="Net revenue" value={money(dashboard.netRevenueMinor, dashboard.currency)} detail={`Gross ${money(dashboard.grossRevenueMinor, dashboard.currency)}`} icon={<Ticket size={22} />} />
        <MetricCard label="Confirmed bookings" value={dashboard.confirmedBookings.toLocaleString()} detail={`${dashboard.heldBookings} currently held`} icon={<Users size={22} />} />
        <MetricCard label="Delivery queue" value={dashboard.pendingDeliveries.toLocaleString()} detail={dashboard.pendingDeliveries === 0 ? "All events delivered" : "Pending or retrying events"} icon={<Clock size={22} />} />
      </section>

      <section className="mt-5 grid gap-5 lg:grid-cols-2">
        <article className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
          <div className="flex items-end justify-between gap-4"><div><p className="text-xs font-bold uppercase tracking-wider text-stone-500">Capacity</p><h2 className="mt-1 text-xl font-bold">Segment occupancy</h2></div><strong>{dashboard.segmentUtilizationPercent.toFixed(2)}%</strong></div>
          <div className="mt-6 h-4 overflow-hidden rounded-full bg-stone-100" role="progressbar" aria-label="Segment utilization" aria-valuenow={dashboard.segmentUtilizationPercent} aria-valuemin={0} aria-valuemax={100}><div className="h-full rounded-full bg-[#851e2e] transition-all" style={{ width: `${Math.min((dashboard.occupiedSeatSegments / capacity) * 100, 100)}%` }} /></div>
          <div className="mt-5 grid grid-cols-2 gap-4 border-t border-stone-100 pt-5"><div><span className="text-xs text-stone-500">Occupied segments</span><strong className="mt-1 block text-xl">{dashboard.occupiedSeatSegments.toLocaleString()}</strong></div><div><span className="text-xs text-stone-500">Sellable segments</span><strong className="mt-1 block text-xl">{dashboard.sellableSeatSegments.toLocaleString()}</strong></div></div>
        </article>

        <article className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm">
          <p className="text-xs font-bold uppercase tracking-wider text-stone-500">Booking health</p><h2 className="mt-1 text-xl font-bold">Status distribution</h2>
          <div className="mt-6 flex h-4 overflow-hidden rounded-full bg-stone-100" aria-label="Booking status distribution"><div className="bg-emerald-600" style={{ width: `${(dashboard.confirmedBookings / bookingTotal) * 100}%` }} /><div className="bg-amber-400" style={{ width: `${(dashboard.heldBookings / bookingTotal) * 100}%` }} /><div className="bg-stone-400" style={{ width: `${(dashboard.cancelledBookings / bookingTotal) * 100}%` }} /></div>
          <dl className="mt-5 grid grid-cols-3 gap-3 text-center"><div className="rounded-xl bg-emerald-50 p-3"><dt className="text-xs text-emerald-700">Confirmed</dt><dd className="mt-1 text-xl font-bold">{dashboard.confirmedBookings}</dd></div><div className="rounded-xl bg-amber-50 p-3"><dt className="text-xs text-amber-700">Held</dt><dd className="mt-1 text-xl font-bold">{dashboard.heldBookings}</dd></div><div className="rounded-xl bg-stone-100 p-3"><dt className="text-xs text-stone-600">Cancelled</dt><dd className="mt-1 text-xl font-bold">{dashboard.cancelledBookings}</dd></div></dl>
        </article>

        <article className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm lg:col-span-2">
          <p className="text-xs font-bold uppercase tracking-wider text-stone-500">Financial reconciliation</p><h2 className="mt-1 text-xl font-bold">Revenue and refunds</h2>
          <div className="mt-6 grid gap-4 sm:grid-cols-3"><div className="rounded-xl border border-stone-200 p-4"><span className="text-xs text-stone-500">Gross captured</span><strong className="mt-1 block text-2xl">{money(dashboard.grossRevenueMinor, dashboard.currency)}</strong></div><div className="rounded-xl border border-stone-200 p-4"><span className="text-xs text-stone-500">Refunded</span><strong className="mt-1 block text-2xl text-amber-700">{money(dashboard.refundedMinor, dashboard.currency)}</strong></div><div className="rounded-xl bg-[#540d17] p-4 text-white"><span className="text-xs text-white/70">Net revenue</span><strong className="mt-1 block text-2xl">{money(dashboard.netRevenueMinor, dashboard.currency)}</strong></div></div>
        </article>
      </section>
    </>
  );
}

export function AdminDashboardView() {
  const [adminKey, setAdminKey] = useState(() => sessionStorage.getItem(SESSION_KEY) ?? "");
  const [date, setDate] = useState(colomboDate);
  const [routeId, setRouteId] = useState("");
  const [runId, setRunId] = useState("");
  const routes = useQuery({ queryKey: ["admin-routes"], queryFn: getRoutes, enabled: Boolean(adminKey) });
  const routeItems: Route[] = useMemo(() => routes.data?.items ?? [], [routes.data]);
  useEffect(() => { if (!routeId && routeItems[0]) setRouteId(routeItems[0].id); }, [routeId, routeItems]);
  const runs = useQuery({ queryKey: ["admin-runs", routeId, date], queryFn: () => getTrainRuns(routeId, date), enabled: Boolean(adminKey && routeId && date) });
  const runItems: TrainRun[] = useMemo(() => runs.data?.items ?? [], [runs.data]);
  useEffect(() => { setRunId((current) => runItems.some((run) => run.id === current) ? current : (runItems[0]?.id ?? "")); }, [runItems]);
  const dashboard = useQuery({ queryKey: ["admin-dashboard", runId], queryFn: () => getAdminDashboard(runId, adminKey), enabled: Boolean(adminKey && runId), refetchInterval: 30_000, retry: false });
  const unauthorized = axios.isAxiosError(dashboard.error) && dashboard.error.response?.status === 401;
  useEffect(() => { if (unauthorized) { sessionStorage.removeItem(SESSION_KEY); setAdminKey(""); } }, [unauthorized]);
  const selectedRun = useMemo(() => runItems.find((run) => run.id === runId), [runItems, runId]);

  if (!adminKey) return <Login onLogin={(key) => { sessionStorage.setItem(SESSION_KEY, key); setAdminKey(key); }} />;
  return (
    <div className="min-h-screen bg-stone-100 text-stone-900">
      <header className="bg-[#540d17] px-5 py-5 text-white"><div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4"><div className="flex items-center gap-3"><span className="rounded-xl bg-white/10 p-2"><TrainFront size={24} /></span><div><p className="text-xs font-bold uppercase tracking-[0.18em] text-amber-200">Lanka Rail Reserve</p><h1 className="text-xl font-bold text-white">Operations control</h1></div></div><div className="flex gap-2"><Link className="secondary-button h-10 bg-white/10 text-white border-white/20 hover:text-stone-900" to="/">Passenger site</Link><Button className="h-10" variant="secondary" onClick={() => { sessionStorage.removeItem(SESSION_KEY); setAdminKey(""); }}>Sign out</Button></div></div></header>
      <main className="mx-auto max-w-7xl px-5 py-8">
        <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm"><div className="grid items-end gap-4 md:grid-cols-[1fr_1fr_1.4fr_auto]"><label className="field"><span>Service date</span><input type="date" value={date} onChange={(event) => setDate(event.target.value)} /></label><label className="field"><span>Route</span><select value={routeId} onChange={(event) => setRouteId(event.target.value)} disabled={routes.isLoading}>{routeItems.map((route) => <option key={route.id} value={route.id}>{route.name}</option>)}</select></label><label className="field"><span>Train run</span><select value={runId} onChange={(event) => setRunId(event.target.value)} disabled={!runItems.length}>{runItems.map((run) => <option key={run.id} value={run.id}>{new Date(run.departureAt).toLocaleTimeString("en-LK", { hour: "2-digit", minute: "2-digit", timeZone: "Asia/Colombo" })} · {run.status}</option>)}</select></label><Button variant="secondary" onClick={() => void dashboard.refetch()} disabled={!runId || dashboard.isFetching}><RefreshCw size={17} className={dashboard.isFetching ? "animate-spin" : ""} />Refresh</Button></div>{selectedRun && <p className="mt-4 text-xs text-stone-500">Run ID <code>{selectedRun.id}</code> · Automatic refresh every 30 seconds</p>}</section>
        {!runId && !runs.isLoading && <div className="mt-5 rounded-2xl border border-dashed border-stone-300 bg-white p-12 text-center"><TrainFront className="mx-auto text-stone-400" size={34} /><h2 className="mt-4 text-xl font-bold">No scheduled train runs</h2><p className="mt-2 text-sm text-stone-500">Choose another service date or route.</p></div>}
        {(routes.isLoading || runs.isLoading || dashboard.isLoading) && <div className="mt-5 flex items-center justify-center gap-3 rounded-2xl bg-white p-12 text-sm font-semibold text-stone-600"><RefreshCw className="animate-spin" />Loading operations data…</div>}
        {dashboard.error && !unauthorized && <div role="alert" className="mt-5 rounded-2xl border border-red-200 bg-red-50 p-5 text-sm font-semibold text-red-800">Dashboard data could not be loaded. Check the API connection and try again.</div>}
        {dashboard.data && <div className="mt-5"><DashboardContent dashboard={dashboard.data} /></div>}
      </main>
    </div>
  );
}
