import type { GraphQLClient, RequestOptions } from "graphql-request";
import gql from "graphql-tag";
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends " $fragmentName" | "__typename" ? T[P] : never };
type GraphQLClientRequestHeaders = RequestOptions["requestHeaders"];
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string };
  String: { input: string; output: string };
  Boolean: { input: boolean; output: boolean };
  Int: { input: number; output: number };
  Float: { input: number; output: number };
  IdJSONType: { input: any; output: any };
  _DirectiveExtensions: { input: any; output: any };
  dmpCrowdCore_JSON: { input: any; output: any };
  dmpCrowdCore_ObjMap: { input: any; output: any };
  dspCore_JSON: { input: any; output: any };
  dspCore_ObjMap: { input: any; output: any };
  sspCore_JSON: { input: any; output: any };
  sspCore_ObjMap: { input: any; output: any };
};

export type CombinedInteraction = {
  __typename?: "CombinedInteraction";
  content?: Maybe<Scalars["IdJSONType"]["output"]>;
};

export type Query = {
  __typename?: "Query";
  combinedInteractionList?: Maybe<CombinedInteraction>;
  combinedInteractionList1?: Maybe<CombinedInteraction>;
  dmpCrowdCore_interactionconfiglist?: Maybe<Scalars["dmpCrowdCore_JSON"]["output"]>;
  dspCore_users?: Maybe<Scalars["dspCore_JSON"]["output"]>;
  sspCore_interactionconfiglist?: Maybe<Scalars["sspCore_JSON"]["output"]>;
};

export enum DmpCrowdCore_HttpMethod {
  Connect = "CONNECT",
  Delete = "DELETE",
  Get = "GET",
  Head = "HEAD",
  Options = "OPTIONS",
  Patch = "PATCH",
  Post = "POST",
  Put = "PUT",
  Trace = "TRACE"
}

export enum DspCore_HttpMethod {
  Connect = "CONNECT",
  Delete = "DELETE",
  Get = "GET",
  Head = "HEAD",
  Options = "OPTIONS",
  Patch = "PATCH",
  Post = "POST",
  Put = "PUT",
  Trace = "TRACE"
}

export enum SspCore_HttpMethod {
  Connect = "CONNECT",
  Delete = "DELETE",
  Get = "GET",
  Head = "HEAD",
  Options = "OPTIONS",
  Patch = "PATCH",
  Post = "POST",
  Put = "PUT",
  Trace = "TRACE"
}

export type CombinedInteractionList_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type CombinedInteractionList_QueryQuery = {
  __typename?: "Query";
  combinedInteractionList?: { __typename?: "CombinedInteraction"; content?: any | null } | null;
};

export type CombinedInteractionList1_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type CombinedInteractionList1_QueryQuery = {
  __typename?: "Query";
  combinedInteractionList1?: { __typename?: "CombinedInteraction"; content?: any | null } | null;
};

export type DmpCrowdCore_Interactionconfiglist_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type DmpCrowdCore_Interactionconfiglist_QueryQuery = {
  __typename?: "Query";
  dmpCrowdCore_interactionconfiglist?: any | null;
};

export type DspCore_Users_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type DspCore_Users_QueryQuery = { __typename?: "Query"; dspCore_users?: any | null };

export type SspCore_Interactionconfiglist_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type SspCore_Interactionconfiglist_QueryQuery = {
  __typename?: "Query";
  sspCore_interactionconfiglist?: any | null;
};

export const CombinedInteractionList_QueryDocument = gql`
  query combinedInteractionList_query {
    combinedInteractionList {
      content
    }
  }
`;
export const CombinedInteractionList1_QueryDocument = gql`
  query combinedInteractionList1_query {
    combinedInteractionList1 {
      content
    }
  }
`;
export const DmpCrowdCore_Interactionconfiglist_QueryDocument = gql`
  query dmpCrowdCore_interactionconfiglist_query {
    dmpCrowdCore_interactionconfiglist
  }
`;
export const DspCore_Users_QueryDocument = gql`
  query dspCore_users_query {
    dspCore_users
  }
`;
export const SspCore_Interactionconfiglist_QueryDocument = gql`
  query sspCore_interactionconfiglist_query {
    sspCore_interactionconfiglist
  }
`;

export type SdkFunctionWrapper = <T>(
  action: (requestHeaders?: Record<string, string>) => Promise<T>,
  operationName: string,
  operationType?: string,
  variables?: any
) => Promise<T>;

const defaultWrapper: SdkFunctionWrapper = (action, _operationName, _operationType, _variables) => action();

export function getSdk(client: GraphQLClient, withWrapper: SdkFunctionWrapper = defaultWrapper) {
  return {
    combinedInteractionList_query(
      variables?: CombinedInteractionList_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<CombinedInteractionList_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<CombinedInteractionList_QueryQuery>({
            document: CombinedInteractionList_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "combinedInteractionList_query",
        "query",
        variables
      );
    },
    combinedInteractionList1_query(
      variables?: CombinedInteractionList1_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<CombinedInteractionList1_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<CombinedInteractionList1_QueryQuery>({
            document: CombinedInteractionList1_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "combinedInteractionList1_query",
        "query",
        variables
      );
    },
    dmpCrowdCore_interactionconfiglist_query(
      variables?: DmpCrowdCore_Interactionconfiglist_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<DmpCrowdCore_Interactionconfiglist_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<DmpCrowdCore_Interactionconfiglist_QueryQuery>({
            document: DmpCrowdCore_Interactionconfiglist_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "dmpCrowdCore_interactionconfiglist_query",
        "query",
        variables
      );
    },
    dspCore_users_query(
      variables?: DspCore_Users_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<DspCore_Users_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<DspCore_Users_QueryQuery>({
            document: DspCore_Users_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "dspCore_users_query",
        "query",
        variables
      );
    },
    sspCore_interactionconfiglist_query(
      variables?: SspCore_Interactionconfiglist_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<SspCore_Interactionconfiglist_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<SspCore_Interactionconfiglist_QueryQuery>({
            document: SspCore_Interactionconfiglist_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "sspCore_interactionconfiglist_query",
        "query",
        variables
      );
    }
  };
}
export type Sdk = ReturnType<typeof getSdk>;
