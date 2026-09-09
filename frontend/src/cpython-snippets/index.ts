import listResizeSrc from "./list_resize@df79316.c?raw";
import ins1Src from "./ins1@df79316.c?raw";
import pyListAppendSrc from "./PyList_Append@df79316.c?raw";

export const CPYTHON_SHA = "df793163d5821791d4e7caf88885a2c11a107986";
export const CPYTHON_VERSION = "3.14.2";

export const snippets: Record<string, string> = {
  "list_resize:growth-formula": listResizeSrc,
  "list_resize:shrink-check": listResizeSrc,
  "list_resize:memcpy": listResizeSrc,
  "list_resize:overflow": listResizeSrc,
  "list_insert:ins1": ins1Src,
  "list_append:PyList_Append": pyListAppendSrc,
  "list_pop:PyList_Pop": listResizeSrc,
  "list_extend:PyList_Extend": listResizeSrc,
};
