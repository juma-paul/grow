/* CPython 3.14.2 — Objects/listobject.c
 * Commit: df793163d5821791d4e7caf88885a2c11a107986
 * https://github.com/python/cpython/blob/v3.14.2/Objects/listobject.c#L475-L501
 */
static int
ins1(PyListObject *self, Py_ssize_t where, PyObject *v)
{
    Py_ssize_t i, n = Py_SIZE(self);
    PyObject **items;
    if (v == NULL) {
        PyErr_BadInternalCall();
        return -1;
    }

    assert((size_t)n + 1 < PY_SSIZE_T_MAX);
    if (list_resize(self, n+1) < 0)
        return -1;

    if (where < 0) {
        where += n;
        if (where < 0)
            where = 0;
    }
    if (where > n)
        where = n;
    items = self->ob_item;
    for (i = n; --i >= where; )
        FT_ATOMIC_STORE_PTR_RELAXED(items[i+1], items[i]);
    FT_ATOMIC_STORE_PTR_RELEASE(items[where], Py_NewRef(v));
    return 0;
}
