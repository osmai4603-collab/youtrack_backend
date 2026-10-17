# مخطط علاقات الصلاحيات (Permission Relationships)

رسم بياني للعلاقات بين الصلاحيات الـ53، مشتقّ من قائمتَي
`dependents` و `implies` في `lib/core/enums/permission_enum.dart`
وجداول `permission_dependencies` في `migrations/000016_rbac_seed.up.sql`.

## الاتجاه (Legend)

> `A --> B`  تعني أن الصلاحية **B** تعتمد على الصلاحية **A**
> (أي A هي الشرط المسبق / prerequisite لـ B).

- العلاقة بين `dependents` و `implies` علاقة عكسية مزدوجة:
  إذا كان `X.implies` يحتوي `Y` فهذا مطابق لوجود `X` في `Y.dependents`.

## المخطط

```mermaid
flowchart LR
    %% ─────────────── Project ───────────────
    subgraph Project["Project"]
        read_project_basic["read-project-basic"]
        create_project["create-project"]
        read_project_full["read-project-full"]
        update_project["update-project"]
        delete_project["delete-project"]
    end

    %% ─────────────── Organization ───────────────
    subgraph Organization["Organization"]
        read_organization["read-organization"]
        update_organization["update-organization"]
        create_organization["create-organization"]
        delete_organization["delete-organization"]
    end

    %% ─────────────── User Profile ───────────────
    subgraph UserProfile["User Profile"]
        update_self["update-self"]
    end

    %% ─────────────── User ───────────────
    subgraph User["User"]
        read_user_basic["read-user-basic"]
        read_user_details["read-user-details"]
        update_user["update-user"]
        create_user["create-user"]
        delete_user["delete-user"]
    end

    %% ─────────────── System ───────────────
    subgraph System["System"]
        low_level_admin_read["low-level-admin-read"]
        low_level_admin_write["low-level-admin-write"]
    end

    %% ─────────────── Issue ───────────────
    subgraph Issue["Issue"]
        read_issue["read-issue"]
        read_issue_private_fields["read-issue-private-fields"]
        update_issue["update-issue"]
        create_issue["create-issue"]
        delete_issue["delete-issue"]
        link_issues["link-issues"]
        update_issue_private_fields["update-issue-private-fields"]
        apply_commands_silently["apply-commands-silently"]
        view_watchers["view-watchers"]
        update_watchers["update-watchers"]
        view_voters["view-voters"]
    end

    %% ─────────────── Attachment ───────────────
    subgraph Attachment["Attachment"]
        add_attachment["add-attachment"]
        update_attachment["update-attachment"]
        delete_attachment["delete-attachment"]
    end

    %% ─────────────── Comment ───────────────
    subgraph Comment["Comment"]
        create_issue_comment["create-issue-comment"]
        read_issue_comment["read-issue-comment"]
        update_issue_comment["update-issue-comment"]
        delete_issue_comment["delete-issue-comment"]
        update_not_own_issue_comment["update-not-own-issue-comment"]
        delete_not_own_comment["delete-not-own-comment-and-permanent-comment-delete"]
        read_article_comment["read-article-comment"]
        create_article_comment["create-article-comment"]
        update_article_comment["update-article-comment"]
        delete_article_comment["delete-article-comment"]
    end

    %% ─────────────── Visibility ───────────────
    subgraph Visibility["Visibility"]
        override_visibility["override-visibility-restrictions"]
    end

    %% ─────────────── Issue Work Item ───────────────
    subgraph WorkItem["Issue Work Item"]
        read_work_item["read-work-item"]
        update_work_item["update-work-item"]
        update_not_own_work_item["update-not-own-work-item"]
        create_work_item["create-work-item"]
        create_not_own_work_item["create-not-own-work-item"]
    end

    %% ─────────────── Article ───────────────
    subgraph Article["Article"]
        read_article["read-article"]
        create_article["create-article"]
        update_article["update-article"]
        delete_article["delete-article"]
    end

    %% ─────────────── App ───────────────
    subgraph App["App"]
        read_app_content["read-app-content"]
        update_app_content["update-app-content"]
    end

    %% ─────────────── Edges (prerequisite --> dependent) ───────────────
    read_project_basic --> read_project_full
    read_project_basic --> read_issue
    read_project_basic --> read_issue_private_fields
    read_project_basic --> create_issue
    read_project_basic --> view_watchers
    read_project_basic --> view_voters
    read_project_basic --> read_article
    read_project_basic --> create_article
    read_project_basic --> update_article
    read_project_basic --> delete_article
    read_project_basic --> read_article_comment
    read_project_basic --> create_article_comment
    read_project_basic --> update_article_comment
    read_project_basic --> delete_article_comment
    read_project_basic --> update_issue_private_fields
    read_project_basic --> override_visibility
    read_project_basic --> update_project
    read_project_basic --> delete_project

    read_project_full --> update_project
    read_project_full --> delete_project

    read_organization --> update_organization
    read_organization --> delete_organization

    update_self --> update_user

    read_user_basic --> read_user_details
    read_user_basic --> update_user
    read_user_basic --> delete_user

    read_user_details --> update_user
    read_user_details --> delete_user

    low_level_admin_read --> low_level_admin_write

    read_issue_private_fields --> update_issue_private_fields
    read_issue_private_fields --> override_visibility
    update_issue --> update_issue_private_fields

    read_issue_comment --> update_not_own_issue_comment
    read_issue_comment --> delete_not_own_comment

    read_article_comment --> create_article_comment
    read_article_comment --> update_article_comment
    read_article_comment --> delete_article_comment

    read_work_item --> update_not_own_work_item
    update_work_item --> update_not_own_work_item
    create_work_item --> create_not_own_work_item

    read_article --> create_article
    read_article --> update_article
    read_article --> delete_article
    read_article --> read_article_comment
    read_article --> create_article_comment
    read_article --> update_article_comment
    read_article --> delete_article_comment

    read_app_content --> update_app_content
```

## المخطط 2: `read-project-basic` من ناحية Dependents

من قائمة `dependents` الخاصة بـ `ProjectReadProjectBasic` في
`permission_enum.dart` (18 صلاحية تعتمد عليها).

> `read-project-basic --> X` تعني أن الصلاحية **X** تعتمد على `read-project-basic`
> (مطلوبة كمقدمة لها).

```mermaid
flowchart LR
    subgraph Hub["المقدمة (Prerequisite)"]
        rpb["read-project-basic"]
    end

    subgraph G_Project["Dependents - Project"]
        rpf["read-project-full"]
        up["update-project"]
        dp["delete-project"]
    end

    subgraph G_Issue["Dependents - Issue"]
        ri["read-issue"]
        ripf["read-issue-private-fields"]
        ci["create-issue"]
        vw["view-watchers"]
        vv["view-voters"]
        uipf["update-issue-private-fields"]
    end

    subgraph G_Article["Dependents - Article"]
        ra["read-article"]
        ca["create-article"]
        ua["update-article"]
        da["delete-article"]
    end

    subgraph G_Comment["Dependents - Comment"]
        rac["read-article-comment"]
        cac["create-article-comment"]
        uac["update-article-comment"]
        dac["delete-article-comment"]
    end

    subgraph G_Vis["Dependents - Visibility"]
        ovr["override-visibility-restrictions"]
    end

    rpb --> rpf
    rpb --> up
    rpb --> dp
    rpb --> ri
    rpb --> ripf
    rpb --> ci
    rpb --> vw
    rpb --> vv
    rpb --> uipf
    rpb --> ra
    rpb --> ca
    rpb --> ua
    rpb --> da
    rpb --> rac
    rpb --> cac
    rpb --> uac
    rpb --> dac
    rpb --> ovr
```

## المخطط 3: `read-project-basic` من ناحية Implies

`ProjectReadProjectBasic.implies` قائمة **فارغة** (لا تمنح أي صلاحية إضافية).

لكن العكس صحيح: توجد **18 صلاحية `implies` تقوم بمنح `read-project-basic`**
عند منحها (أي `X.implies` يحتوي `read-project-basic`).

> `X -.implies.-> read-project-basic` تعني أن منح الصلاحية **X** يمنح
> تلقائيًا صلاحية `read-project-basic`.

```mermaid
flowchart LR
    subgraph Hub2["الصلاحية الممنوحة"]
        rpb2["read-project-basic"]
    end

    subgraph Imp_Project["Implies - Project"]
        rpf2["read-project-full"]
        up2["update-project"]
        dp2["delete-project"]
    end

    subgraph Imp_Issue["Implies - Issue"]
        ri2["read-issue"]
        ripf2["read-issue-private-fields"]
        ci2["create-issue"]
        vw2["view-watchers"]
        vv2["view-voters"]
        uipf2["update-issue-private-fields"]
    end

    subgraph Imp_Article["Implies - Article"]
        ra2["read-article"]
        ca2["create-article"]
        ua2["update-article"]
        da2["delete-article"]
    end

    subgraph Imp_Comment["Implies - Comment"]
        rac2["read-article-comment"]
        cac2["create-article-comment"]
        uac2["update-article-comment"]
        dac2["delete-article-comment"]
    end

    subgraph Imp_Vis["Implies - Visibility"]
        ovr2["override-visibility-restrictions"]
    end

    rpf2 -.implies.-> rpb2
    up2 -.implies.-> rpb2
    dp2 -.implies.-> rpb2
    ri2 -.implies.-> rpb2
    ripf2 -.implies.-> rpb2
    ci2 -.implies.-> rpb2
    vw2 -.implies.-> rpb2
    vv2 -.implies.-> rpb2
    uipf2 -.implies.-> rpb2
    ra2 -.implies.-> rpb2
    ca2 -.implies.-> rpb2
    ua2 -.implies.-> rpb2
    da2 -.implies.-> rpb2
    rac2 -.implies.-> rpb2
    cac2 -.implies.-> rpb2
    uac2 -.implies.-> rpb2
    dac2 -.implies.-> rpb2
    ovr2 -.implies.-> rpb2
```

## المخطط 4: `read-project-full`

صلاحية **مشتركة** في الحالتين: لها `dependents` ولها `implies`.

> `read-project-full --> X` تعني أن X تعتمد عليها، وأن
> `read-project-full -.implies.-> Y` تعني أنها تمنح Y عند منحها.

```mermaid
flowchart LR
    subgraph Hub_A["الصلاحية المدروسة"]
        h_rpf["read-project-full"]
    end

    subgraph Dep_A["Dependents (تعتمد عليها)"]
        d_up["update-project"]
        d_dp["delete-project"]
    end

    subgraph Imp_A["Implies (تمنحها)"]
        i_rpb["read-project-basic"]
    end

    h_rpf --> d_up
    h_rpf --> d_dp
    h_rpf -.implies.-> i_rpb
```

## المخطط 5: `read-user-details`

صلاحية **مشتركة** في الحالتين: لها `dependents` ولها `implies`.

> `read-user-details --> X` تعني أن X تعتمد عليها، وأن
> `read-user-details -.implies.-> Y` تعني أنها تمنح Y عند منحها.

```mermaid
flowchart LR
    subgraph Hub_B["الصلاحية المدروسة"]
        h_rud["read-user-details"]
    end

    subgraph Dep_B["Dependents (تعتمد عليها)"]
        d_upd["update-user"]
        d_dud["delete-user"]
    end

    subgraph Imp_B["Implies (تمنحها)"]
        i_rub["read-user-basic"]
    end

    h_rud --> d_upd
    h_rud --> d_dud
    h_rud -.implies.-> i_rub
```

## المخطط 6: `read-issue-private-fields`

صلاحية **مشتركة** في الحالتين: لها `dependents` ولها `implies`.

> `read-issue-private-fields --> X` تعني أن X تعتمد عليها، وأن
> `read-issue-private-fields -.implies.-> Y` تعني أنها تمنح Y عند منحها.

```mermaid
flowchart LR
    subgraph Hub_C["الصلاحية المدروسة"]
        h_ripf["read-issue-private-fields"]
    end

    subgraph Dep_C["Dependents (تعتمد عليها)"]
        d_uipf["update-issue-private-fields"]
        d_ovr["override-visibility-restrictions"]
    end

    subgraph Imp_C["Implies (تمنحها)"]
        i_rpb2["read-project-basic"]
    end

    h_ripf --> d_uipf
    h_ripf --> d_ovr
    h_ripf -.implies.-> i_rpb2
```

## المخطط 7: `read-article-comment`

صلاحية **مشتركة** في الحالتين: لها `dependents` ولها `implies`.

> `read-article-comment --> X` تعني أن X تعتمد عليها، وأن
> `read-article-comment -.implies.-> Y` تعني أنها تمنح Y عند منحها.

```mermaid
flowchart LR
    subgraph Hub_D["الصلاحية المدروسة"]
        h_rac["read-article-comment"]
    end

    subgraph Dep_D["Dependents (تعتمد عليها)"]
        d_cac["create-article-comment"]
        d_uac["update-article-comment"]
        d_dac["delete-article-comment"]
    end

    subgraph Imp_D["Implies (تمنحها)"]
        i_ra["read-article"]
        i_rpb3["read-project-basic"]
    end

    h_rac --> d_cac
    h_rac --> d_uac
    h_rac --> d_dac
    h_rac -.implies.-> i_ra
    h_rac -.implies.-> i_rpb3
```

## المخطط 8: `read-article`

صلاحية **مشتركة** في الحالتين: لها `dependents` ولها `implies`.

> `read-article --> X` تعني أن X تعتمد عليها، وأن
> `read-article -.implies.-> Y` تعني أنها تمنح Y عند منحها.

```mermaid
flowchart LR
    subgraph Hub_E["الصلاحية المدروسة"]
        h_ra["read-article"]
    end

    subgraph Dep_E["Dependents (تعتمد عليها)"]
        d_cra["create-article"]
        d_ura["update-article"]
        d_dra["delete-article"]
        d_rac2["read-article-comment"]
        d_cac2["create-article-comment"]
        d_uac2["update-article-comment"]
        d_dac2["delete-article-comment"]
    end

    subgraph Imp_E["Implies (تمنحها)"]
        i_rpb4["read-project-basic"]
    end

    h_ra --> d_cra
    h_ra --> d_ura
    h_ra --> d_dra
    h_ra --> d_rac2
    h_ra --> d_cac2
    h_ra --> d_uac2
    h_ra --> d_dac2
    h_ra -.implies.-> i_rpb4
```

## ملاحظات

- الصلاحيات **المشتركة** بين `dependents` و `implies` (مخططات 4-8):
  `read-project-full`, `read-user-details`, `read-issue-private-fields`,
  `read-article-comment`, `read-article`.
- **`read-project-basic`** هي العقدة الجذرية الأكثر تشعبًا (18 صلة إخراج): معظم
  صلاحيات المشروع/القضية/المقالة تعتمد عليها.
- الصلاحيات **بدون أي صلة اعتماد** (أوراق/حرة): `create-project`,
  `create-organization`, `create-user`, `delete-issue`, `link-issues`,
  `apply-commands-silently`, `update-watchers`, `add-attachment`,
  `update-attachment`, `delete-attachment`, `create-issue-comment`,
  `update-issue-comment`, `delete-issue-comment`.
- `is_global = TRUE` للصلاحيات غير المرتبطة بمشروع (User/System/Org/App)،
  كما هو موثّق في `migrations/000016_rbac_seed.up.sql`.
