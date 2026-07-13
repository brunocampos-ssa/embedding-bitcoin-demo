import{createContext,useContext,useMemo,useState,type Dispatch,type ReactNode,type SetStateAction}from'react';import{messages,type Locale,type MessageKey}from'./translations';
const fallback:Locale='en-US';const initial=():Locale=>navigator.language.toLowerCase().startsWith('pt')?'pt-BR':fallback;
type I18nValue={locale:Locale;setLocale:Dispatch<SetStateAction<Locale>>;t:(k:MessageKey)=>string};
const C=createContext<I18nValue>({locale:fallback,setLocale:()=>{},t:(k:MessageKey)=>messages[fallback][k]});
export function I18nProvider({children}:{children:ReactNode}){const[locale,setLocale]=useState<Locale>(initial);const value=useMemo(()=>({locale,setLocale,t:(k:MessageKey)=>messages[locale][k]}),[locale]);return <C.Provider value={value}>{children}</C.Provider>}
export const useI18n=()=>useContext(C);
