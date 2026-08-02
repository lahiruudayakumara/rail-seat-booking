import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export function getTomorrowDate(): string {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  return d.toISOString().slice(0, 10);
}

export interface SearchState {
  routeId: string;
  originId: string;
  destinationId: string;
  date: string;
  searched: boolean;
}

const initialState: SearchState = {
  routeId: "",
  originId: "",
  destinationId: "",
  date: getTomorrowDate(),
  searched: false,
};

export const searchSlice = createSlice({
  name: "search",
  initialState,
  reducers: {
    setRouteId: (state, action: PayloadAction<string>) => {
      if (state.routeId === action.payload) return;
      state.routeId = action.payload;
      state.originId = "";
      state.destinationId = "";
      state.searched = false;
    },
    setOriginId: (state, action: PayloadAction<string>) => {
      state.originId = action.payload;
    },
    setDestinationId: (state, action: PayloadAction<string>) => {
      state.destinationId = action.payload;
    },
    setDate: (state, action: PayloadAction<string>) => {
      state.date = action.payload;
    },
    setSearched: (state, action: PayloadAction<boolean>) => {
      state.searched = action.payload;
    },
    resetSearch: (state) => {
      state.searched = false;
    },
  },
});

export const {
  setRouteId,
  setOriginId,
  setDestinationId,
  setDate,
  setSearched,
  resetSearch,
} = searchSlice.actions;

export default searchSlice.reducer;
