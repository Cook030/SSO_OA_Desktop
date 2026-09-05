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
  mhSso2_JSON: { input: any; output: any };
  mhSso2_ObjMap: { input: any; output: any };
};

export type CombinedInteraction = {
  __typename?: "CombinedInteraction";
  content?: Maybe<Scalars["IdJSONType"]["output"]>;
};

export type Mutation = {
  __typename?: "Mutation";
  mhSso2_authLogin?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
  mhSso2_authLogout?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
  mhSso2_authRefresh?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
  mhSso2_authRegister?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
  mhSso2_saveMenuConfig?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
};

export type MutationMhSso2_AuthLoginArgs = {
  input?: InputMaybe<MhSso2_MutationInput_AuthLogin_Input_Input>;
};

export type MutationMhSso2_AuthLogoutArgs = {
  input?: InputMaybe<MhSso2_MutationInput_AuthLogout_Input_Input>;
};

export type MutationMhSso2_AuthRefreshArgs = {
  input?: InputMaybe<MhSso2_MutationInput_AuthRefresh_Input_Input>;
};

export type MutationMhSso2_AuthRegisterArgs = {
  input?: InputMaybe<MhSso2_MutationInput_AuthRegister_Input_Input>;
};

export type MutationMhSso2_SaveMenuConfigArgs = {
  input?: InputMaybe<Scalars["mhSso2_JSON"]["input"]>;
};

export type Query = {
  __typename?: "Query";
  combinedInteractionList?: Maybe<CombinedInteraction>;
  combinedInteractionList1?: Maybe<CombinedInteraction>;
  mhSso2_authMe?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
  mhSso2_getMenuConfig?: Maybe<Scalars["mhSso2_JSON"]["output"]>;
};

export enum MhSso2_HttpMethod {
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

export type MhSso2_MutationInput_AuthLogin_Input_Input = {
  account: Scalars["String"]["input"];
  password: Scalars["String"]["input"];
};

export type MhSso2_MutationInput_AuthLogout_Input_Input = {
  refreshToken?: InputMaybe<Scalars["String"]["input"]>;
};

export type MhSso2_MutationInput_AuthRefresh_Input_Input = {
  refreshToken?: InputMaybe<Scalars["String"]["input"]>;
};

export type MhSso2_MutationInput_AuthRegister_Input_Input = {
  confirmPassword: Scalars["String"]["input"];
  email?: InputMaybe<Scalars["String"]["input"]>;
  mobile?: InputMaybe<Scalars["String"]["input"]>;
  nickname?: InputMaybe<Scalars["String"]["input"]>;
  password: Scalars["String"]["input"];
  username: Scalars["String"]["input"];
};

export type MhSso2_AuthLogin_MutationMutationVariables = Exact<{
  input?: InputMaybe<MhSso2_MutationInput_AuthLogin_Input_Input>;
}>;

export type MhSso2_AuthLogin_MutationMutation = { __typename?: "Mutation"; mhSso2_authLogin?: any | null };

export type MhSso2_AuthLogout_MutationMutationVariables = Exact<{
  input?: InputMaybe<MhSso2_MutationInput_AuthLogout_Input_Input>;
}>;

export type MhSso2_AuthLogout_MutationMutation = { __typename?: "Mutation"; mhSso2_authLogout?: any | null };

export type MhSso2_AuthRefresh_MutationMutationVariables = Exact<{
  input?: InputMaybe<MhSso2_MutationInput_AuthRefresh_Input_Input>;
}>;

export type MhSso2_AuthRefresh_MutationMutation = { __typename?: "Mutation"; mhSso2_authRefresh?: any | null };

export type MhSso2_AuthRegister_MutationMutationVariables = Exact<{
  input?: InputMaybe<MhSso2_MutationInput_AuthRegister_Input_Input>;
}>;

export type MhSso2_AuthRegister_MutationMutation = { __typename?: "Mutation"; mhSso2_authRegister?: any | null };

export type MhSso2_SaveMenuConfig_MutationMutationVariables = Exact<{
  input?: InputMaybe<Scalars["mhSso2_JSON"]["input"]>;
}>;

export type MhSso2_SaveMenuConfig_MutationMutation = { __typename?: "Mutation"; mhSso2_saveMenuConfig?: any | null };

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

export type MhSso2_AuthMe_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type MhSso2_AuthMe_QueryQuery = { __typename?: "Query"; mhSso2_authMe?: any | null };

export type MhSso2_GetMenuConfig_QueryQueryVariables = Exact<{ [key: string]: never }>;

export type MhSso2_GetMenuConfig_QueryQuery = { __typename?: "Query"; mhSso2_getMenuConfig?: any | null };

export const MhSso2_AuthLogin_MutationDocument = gql`
  mutation mhSso2_authLogin_mutation($input: mhSso2_mutationInput_authLogin_input_Input) {
    mhSso2_authLogin(input: $input)
  }
`;
export const MhSso2_AuthLogout_MutationDocument = gql`
  mutation mhSso2_authLogout_mutation($input: mhSso2_mutationInput_authLogout_input_Input) {
    mhSso2_authLogout(input: $input)
  }
`;
export const MhSso2_AuthRefresh_MutationDocument = gql`
  mutation mhSso2_authRefresh_mutation($input: mhSso2_mutationInput_authRefresh_input_Input) {
    mhSso2_authRefresh(input: $input)
  }
`;
export const MhSso2_AuthRegister_MutationDocument = gql`
  mutation mhSso2_authRegister_mutation($input: mhSso2_mutationInput_authRegister_input_Input) {
    mhSso2_authRegister(input: $input)
  }
`;
export const MhSso2_SaveMenuConfig_MutationDocument = gql`
  mutation mhSso2_saveMenuConfig_mutation($input: mhSso2_JSON) {
    mhSso2_saveMenuConfig(input: $input)
  }
`;
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
export const MhSso2_AuthMe_QueryDocument = gql`
  query mhSso2_authMe_query {
    mhSso2_authMe
  }
`;
export const MhSso2_GetMenuConfig_QueryDocument = gql`
  query mhSso2_getMenuConfig_query {
    mhSso2_getMenuConfig
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
    mhSso2_authLogin_mutation(
      variables?: MhSso2_AuthLogin_MutationMutationVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_AuthLogin_MutationMutation> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_AuthLogin_MutationMutation>({
            document: MhSso2_AuthLogin_MutationDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_authLogin_mutation",
        "mutation",
        variables
      );
    },
    mhSso2_authLogout_mutation(
      variables?: MhSso2_AuthLogout_MutationMutationVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_AuthLogout_MutationMutation> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_AuthLogout_MutationMutation>({
            document: MhSso2_AuthLogout_MutationDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_authLogout_mutation",
        "mutation",
        variables
      );
    },
    mhSso2_authRefresh_mutation(
      variables?: MhSso2_AuthRefresh_MutationMutationVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_AuthRefresh_MutationMutation> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_AuthRefresh_MutationMutation>({
            document: MhSso2_AuthRefresh_MutationDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_authRefresh_mutation",
        "mutation",
        variables
      );
    },
    mhSso2_authRegister_mutation(
      variables?: MhSso2_AuthRegister_MutationMutationVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_AuthRegister_MutationMutation> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_AuthRegister_MutationMutation>({
            document: MhSso2_AuthRegister_MutationDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_authRegister_mutation",
        "mutation",
        variables
      );
    },
    mhSso2_saveMenuConfig_mutation(
      variables?: MhSso2_SaveMenuConfig_MutationMutationVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_SaveMenuConfig_MutationMutation> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_SaveMenuConfig_MutationMutation>({
            document: MhSso2_SaveMenuConfig_MutationDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_saveMenuConfig_mutation",
        "mutation",
        variables
      );
    },
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
    mhSso2_authMe_query(
      variables?: MhSso2_AuthMe_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_AuthMe_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_AuthMe_QueryQuery>({
            document: MhSso2_AuthMe_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_authMe_query",
        "query",
        variables
      );
    },
    mhSso2_getMenuConfig_query(
      variables?: MhSso2_GetMenuConfig_QueryQueryVariables,
      requestHeaders?: GraphQLClientRequestHeaders,
      signal?: RequestInit["signal"]
    ): Promise<MhSso2_GetMenuConfig_QueryQuery> {
      return withWrapper(
        wrappedRequestHeaders =>
          client.request<MhSso2_GetMenuConfig_QueryQuery>({
            document: MhSso2_GetMenuConfig_QueryDocument,
            variables,
            requestHeaders: { ...requestHeaders, ...wrappedRequestHeaders },
            signal
          }),
        "mhSso2_getMenuConfig_query",
        "query",
        variables
      );
    }
  };
}
export type Sdk = ReturnType<typeof getSdk>;
