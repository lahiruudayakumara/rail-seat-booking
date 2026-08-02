import { configureStore } from "@reduxjs/toolkit";
import { type TypedUseSelectorHook, useDispatch, useSelector } from "react-redux";
import bookingReducer from "./slices/booking-slice";
import searchReducer from "./slices/search-slice";
import uiReducer from "./slices/ui-slice";

export const store = configureStore({
  reducer: {
    search: searchReducer,
    booking: bookingReducer,
    ui: uiReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;
