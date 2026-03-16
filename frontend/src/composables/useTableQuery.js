import { ref, watch } from "vue"
import { useRouter, useRoute } from "vue-router"

export function useTableQuery(defaults = {}) {
    const router = useRouter()
    const route = useRoute()

    const q = route.query

    const search = ref(q.search || defaults.search || "")
    const filterStatus = ref(q.status || defaults.status || "")
    const sortBy = ref(q.sort_by || defaults.sortBy || "created_at")
    const sortDir = ref(q.sort_dir || defaults.sortDir || "desc")
    const page = ref(Number(q.page) || defaults.page || 1)
    const pageSize = ref(Number(q.page_size) || defaults.pageSize || 20)

    function pushQuery() {
        const query = {}
        if (search.value) query.search = search.value
        if (filterStatus.value) query.status = filterStatus.value
        query.sort_by = sortBy.value
        query.sort_dir = sortDir.value
        query.page = String(page.value)
        query.page_size = String(pageSize.value)
        router.replace({ query })
    }

    function buildParams() {
        const params = new URLSearchParams()
        if (search.value) params.set("search", search.value)
        if (filterStatus.value) params.set("status", filterStatus.value)
        params.set("sort_by", sortBy.value)
        params.set("sort_dir", sortDir.value)
        params.set("page", page.value)
        params.set("page_size", pageSize.value)
        return params
    }

    return { search, filterStatus, sortBy, sortDir, page, pageSize, pushQuery, buildParams }
}
