export interface FaqItem {
  id: string;
  questionKey: string;
  answerKey: string;
}

export const FAQ_ITEMS: FaqItem[] = [
  { id: "q1", questionKey: "help.q1", answerKey: "help.a1" },
  { id: "q2", questionKey: "help.q2", answerKey: "help.a2" },
  { id: "q3", questionKey: "help.q3", answerKey: "help.a3" },
];
