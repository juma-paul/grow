/* CPython 3.14.2 — Objects/listobject.c
 * Commit: df793163d5821791d4e7caf88885a2c11a107986
 * https://github.com/python/cpython/blob/v3.14.2/Objects/listobject.c#L518-L544
 */

/* internal, used by _PyList_AppendTakeRef */
int
_PyList_AppendTakeRefListResize(PyListObject *self, PyObject *newitem)
{
    Py_ssize_t len = Py_SIZE(self);
    assert(self->allocated == -1 || self->allocated == len);
    if (list_resize(self, len + 1) < 0) {
        Py_DECREF(newitem);
        return -1;
    }
    FT_ATOMIC_STORE_PTR_RELEASE(self->ob_item[len], newitem);
    return 0;
}

int
PyList_Append(PyObject *op, PyObject *newitem)
{
    if (PyList_Check(op) && (newitem != NULL)) {
        int ret;
        Py_BEGIN_CRITICAL_SECTION(op);
        ret = _PyList_AppendTakeRef((PyListObject *)op, Py_NewRef(newitem));
        Py_END_CRITICAL_SECTION();
        return ret;
    }
    PyErr_BadInternalCall();
    return -1;
}
