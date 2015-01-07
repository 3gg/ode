package ode

// #include <ode/ode.h>
// extern int callTriCallback(dGeomID mesh, dGeomID other, int index);
// extern int callTriRayCallback(dGeomID mesh, dGeomID ray, int index, dReal u, dReal v);
import "C"

import (
	"unsafe"
)

var (
	triCallbacks    = map[TriMesh]TriCallback{}
	triRayCallbacks = map[TriMesh]TriRayCallback{}
)

// Important note: The TriMeshData copies the vertex/index and normal data.
// If Destroy isn't called a memory leak will occur.

// TriMeshData represents triangle mesh data.
type TriMeshData uintptr

type triMeshDataCopy struct {
	verts   unsafe.Pointer
	nverts  C.int
	vstride C.int
	tris    unsafe.Pointer
	ntris   C.int
	tstride C.int
}

func newTriMeshDataCopy(verts []float64, tris []uint32) triMeshDataCopy {
	t := triMeshDataCopy{}

	t.verts = C.malloc(C.size_t(len(verts) * 8))
	C.memcpy(t.verts, unsafe.Pointer(&verts[0]), C.size_t(len(verts)*8))
	t.nverts = C.int(len(verts))
	t.vstride = C.int(8 * 4)

	t.tris = C.malloc(C.size_t(len(tris) * 4))
	t.ntris = C.int(len(tris))
	t.tstride = C.int(3 * 4)
	C.memcpy(t.tris, unsafe.Pointer(&tris[0]), C.size_t(len(tris)*4))

	return t
}

func newTriMeshDataCopyDouble(verts []float64, tris []uint32) triMeshDataCopy {
	t := triMeshDataCopy{}

	t.verts = C.malloc(C.size_t(len(verts) * 8))
	C.memcpy(t.verts, unsafe.Pointer(&verts[0]), C.size_t(len(verts)*8))
	t.nverts = C.int(len(verts))
	t.vstride = C.int(8 * 3)

	t.tris = C.malloc(C.size_t(len(tris) * 4))
	t.ntris = C.int(len(tris))
	t.tstride = C.int(3 * 4)
	C.memcpy(t.tris, unsafe.Pointer(&tris[0]), C.size_t(len(tris)*4))

	return t
}

func newTriMeshDataCopySingle(verts []float32, tris []uint32) triMeshDataCopy {
	t := triMeshDataCopy{}

	t.verts = C.malloc(C.size_t(len(verts) * 4))
	C.memcpy(t.verts, unsafe.Pointer(&verts[0]), C.size_t(len(verts)*4))
	t.nverts = C.int(len(verts))
	t.vstride = C.int(4 * 3)

	t.tris = C.malloc(C.size_t(len(tris) * 4))
	t.ntris = C.int(len(tris))
	t.tstride = C.int(3 * 4)
	C.memcpy(t.tris, unsafe.Pointer(&tris[0]), C.size_t(len(tris)*4))

	return t
}

func (t triMeshDataCopy) destroy() {
	if t.verts != nil {
		C.free(t.verts)
		t.verts = nil
	}
	if t.tris != nil {
		C.free(t.tris)
		t.tris = nil
	}
}

var (
	triMeshData map[TriMeshData]triMeshDataCopy = make(map[TriMeshData]triMeshDataCopy, 0)
)

func cToTriMeshData(c C.dTriMeshDataID) TriMeshData {
	return TriMeshData(unsafe.Pointer(c))
}

func (t TriMeshData) c() C.dTriMeshDataID {
	return C.dTriMeshDataID(unsafe.Pointer(t))
}

// NewTriMeshData returns a new TriMeshData instance.
func NewTriMeshData() TriMeshData {
	return cToTriMeshData(C.dGeomTriMeshDataCreate())
}

// Destroy destroys the triangle mesh data.
func (t TriMeshData) Destroy() {
	data, ok := triMeshData[t]
	if ok {
		data.destroy()
		delete(triMeshData, t)
	}
	C.dGeomTriMeshDataDestroy(t.c())
}

// 4 float64 per vert is required here.
func (t TriMeshData) Build(verts []float64, tris []uint32) {
	data, ok := triMeshData[t]
	if ok {
		data.destroy()
	}
	data = newTriMeshDataCopy(verts, tris)
	triMeshData[t] = data
	C.dGeomTriMeshDataBuildSimple(t.c(), (*C.dReal)(data.verts), data.nverts,
		(*C.dTriIndex)(data.tris), data.ntris)
}

// 3 float64 per vert.
func (t TriMeshData) BuildDouble(verts []float64, tris []uint32) {
	data, ok := triMeshData[t]
	if ok {
		data.destroy()
	}
	data = newTriMeshDataCopyDouble(verts, tris)
	triMeshData[t] = data

	C.dGeomTriMeshDataBuildDouble(t.c(),
		data.verts, data.vstride, data.nverts,
		data.tris, data.ntris, data.tstride)
}

// 3 float32 per vert.
func (t TriMeshData) BuildSingle(verts []float32, tris []uint32) {
	data, ok := triMeshData[t]
	if ok {
		data.destroy()
	}
	data = newTriMeshDataCopySingle(verts, tris)
	triMeshData[t] = data

	C.dGeomTriMeshDataBuildSingle(t.c(),
		data.verts, data.vstride, data.nverts,
		data.tris, data.ntris, data.tstride)
}

// TODO: Add support for more complex trimesh data?

// Preprocess preprocesses the triangle mesh data.
func (t TriMeshData) Preprocess() {
	C.dGeomTriMeshDataPreprocess(t.c())
}

// Update updates the triangle mesh data.
func (t TriMeshData) Update() {
	C.dGeomTriMeshDataUpdate(t.c())
}

// TriMesh is a geometry representing a triangle mesh.
type TriMesh struct {
	GeomBase
}

// TriCallback is called to determine whether to collide a triangle with
// another geometry.
type TriCallback func(mesh TriMesh, other Geom, index int) bool

//export triCallback
func triCallback(c C.dGeomID, other C.dGeomID, index C.int) C.int {
	mesh := cToGeom(c).(TriMesh)
	cb, ok := triCallbacks[mesh]
	if !ok {
		return 0
	}
	return C.int(btoi(cb(mesh, cToGeom(other), int(index))))
}

// TriRayCallback is called to determine whether to collide a triangle with a
// ray at a given point.
type TriRayCallback func(mesh TriMesh, ray Ray, index int, u, v float64) bool

//export triRayCallback
func triRayCallback(c C.dGeomID, ray C.dGeomID, index C.int, u, v C.dReal) C.int {
	mesh := cToGeom(c).(TriMesh)
	cb, ok := triRayCallbacks[mesh]
	if !ok {
		return 0
	}
	return C.int(btoi(cb(mesh, cToGeom(ray).(Ray), int(index), float64(u), float64(v))))
}

// SetLastTransform sets the last transform.
func (t TriMesh) SetLastTransform(xform Matrix4) {
	C.dGeomTriMeshSetLastTransform(t.c(), (*C.dReal)(&xform[0][0]))
}

// LastTransform returns the last transform.
func (t TriMesh) LastTransform() Matrix4 {
	xform := NewMatrix4()
	c := C.dGeomTriMeshGetLastTransform(t.c())
	Matrix(xform).fromC(c)
	return xform
}

// SetTriCallback sets the triangle collision callback.
func (t TriMesh) SetTriCallback(cb TriCallback) {
	if cb == nil {
		C.dGeomTriMeshSetCallback(t.c(), nil) // clear callback
		delete(triCallbacks, t)
	} else {
		triCallbacks[t] = cb
		C.dGeomTriMeshSetCallback(t.c(), (*C.dTriCallback)(C.callTriCallback))
	}
}

// TriCallback returns the triangle collision callback.
func (t TriMesh) TriCallback() TriCallback {
	return triCallbacks[t]
}

// SetTriRayCallback sets the triangle/ray collision callback.
func (t TriMesh) SetTriRayCallback(cb TriRayCallback) {
	if cb == nil {
		C.dGeomTriMeshSetCallback(t.c(), nil) // clear callback
		delete(triRayCallbacks, t)
	} else {
		triRayCallbacks[t] = cb
		C.dGeomTriMeshSetRayCallback(t.c(), (*C.dTriRayCallback)(C.callTriRayCallback))
	}
}

// TriRayCallback returns the triangle/ray collision callback.
func (t TriMesh) TriRayCallback() TriRayCallback {
	return triRayCallbacks[t]
}

// SetMeshData sets the mesh data.
func (t TriMesh) SetMeshData(data TriMeshData) {
	C.dGeomTriMeshSetData(t.c(), data.c())
}

// MeshData returns the mesh data.
func (t TriMesh) MeshData() TriMeshData {
	return cToTriMeshData(C.dGeomTriMeshGetData(t.c()))
}

// SetTCEnabled sets whether temporal coherence is enabled for the given
// geometry class.
func (t TriMesh) SetTCEnabled(class int, isEnabled bool) {
	C.dGeomTriMeshEnableTC(t.c(), C.int(class), C.int(btoi(isEnabled)))
}

// TCEnabled returns whether temporal coherence is enabled for the given
// geometry class.
func (t TriMesh) TCEnabled(class int) bool {
	return C.dGeomTriMeshIsTCEnabled(t.c(), C.int(class)) != 0
}

// ClearTCCache clears the temporal coherence cache.
func (t TriMesh) ClearTCCache() {
	C.dGeomTriMeshClearTCCache(t.c())
}

// Triangle returns a triangle in the mesh by index.
func (t TriMesh) Triangle(index int) (Vector3, Vector3, Vector3) {
	c0, c1, c2 := &C.dVector3{}, &C.dVector3{}, &C.dVector3{}
	v0, v1, v2 := NewVector3(), NewVector3(), NewVector3()
	C.dGeomTriMeshGetTriangle(t.c(), C.int(index), c0, c1, c2)
	Vector(v0).fromC(&c0[0])
	Vector(v1).fromC(&c1[0])
	Vector(v2).fromC(&c2[0])
	return v0, v1, v2
}

// Point returns a point on the specified triangle at the given barycentric coordinates.
func (t TriMesh) Point(index int, u, v float64) Vector3 {
	pt := NewVector3()
	C.dGeomTriMeshGetPoint(t.c(), C.int(index), C.dReal(u), C.dReal(v), (*C.dReal)(&pt[0]))
	return pt
}

// TriangleCount returns the number of triangles in the mesh.
func (t TriMesh) TriangleCount() int {
	return int(C.dGeomTriMeshGetTriangleCount(t.c()))
}
