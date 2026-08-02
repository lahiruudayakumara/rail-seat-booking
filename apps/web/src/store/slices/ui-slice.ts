import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export interface UiState {
  notice: string;
}

const initialState: UiState = {
  notice: "",
};

export const uiSlice = createSlice({
  name: "ui",
  initialState,
  reducers: {
    setNotice: (state, action: PayloadAction<string>) => {
      state.notice = action.payload;
    },
    clearNotice: (state) => {
      state.notice = "";
    },
  },
});

export const { setNotice, clearNotice } = uiSlice.actions;

export default uiSlice.reducer;
