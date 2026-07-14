export type BountyState='DRAFT'|'OPEN'|'ASSIGNED'|'SUBMITTED'|'APPROVED'|'PAID'|'CANCELLED'|'EXPIRED';
export interface Submission{id:string;actor:string;evidenceUrl:string;notes:string;state:string}
export interface Bounty{id:string;title:string;description:string;format:string;language:string;rewardSats:number;state:BountyState;submissions?:Submission[]}
export interface Payout{id:string;submissionId:string;state:string;asset:string;rail:string;amountBaseUnits:number;feeBaseUnits:number;destinationType:string;destinationMasked:string;providerPaymentId?:string;expiresAt:string;failureCode?:string}
export interface Health{status:string;paymentProvider:string;demoMode:boolean;authentication:string}
export type DepositRail='lightning'|'bitcoin'|'spark';
export interface Treasury{balanceSats:number;identity:{lightningAddress?:string;sparkAddress?:string};tokenBalances?:Record<string,number>}
export interface DepositQuote{rail:DepositRail;paymentRequest:string;feeSats:number;expiresAt?:string}
