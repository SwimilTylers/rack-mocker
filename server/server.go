package server

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"os"
	"rack-mocker/file"
	"sync"
)

type fileRead struct {
	request *file.FileReadRequest
	respC   chan<- *file.FileReadResponse
}

type fileWrite struct {
	request *file.FileWriteRequest
	respC   chan<- *file.FileWriteResponse
}

type fileOp struct {
	read  *fileRead
	write *fileWrite
	errC  chan error
}

type fileOpPipe chan *fileOp

type Server struct {
	fs fs.FS

	pMu *sync.Mutex

	pipes map[string]fileOpPipe
	stopC chan struct{}
}

func NewService(root string) *Server {
	return &Server{
		fs:    os.DirFS(root),
		pMu:   &sync.Mutex{},
		pipes: map[string]fileOpPipe{},
		stopC: make(chan struct{}),
	}
}

func (s *Server) FileList(ctx context.Context, in *file.FileListRequest) (*file.FileListResponse, error) {
	return nil, errors.New("not supported yet")
}

func (s *Server) FileRead(ctx context.Context, in *file.FileReadRequest) (*file.FileReadResponse, error) {
	p := s.fetchPipe(in.Filepath)

	respC := make(chan *file.FileReadResponse)
	errC := make(chan error)
	p <- &fileOp{read: &fileRead{request: in, respC: respC}, errC: errC}

	select {
	case resp := <-respC:
		return resp, nil
	case err := <-errC:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Server) FileWrite(ctx context.Context, in *file.FileWriteRequest) (*file.FileWriteResponse, error) {
	p := s.fetchPipe(in.Filepath)

	respC := make(chan *file.FileWriteResponse)
	errC := make(chan error)
	p <- &fileOp{write: &fileWrite{request: in, respC: respC}, errC: errC}

	select {
	case resp := <-respC:
		return resp, nil
	case err := <-errC:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *Server) fetchPipe(filepath string) fileOpPipe {
	s.pMu.Lock()
	defer s.pMu.Unlock()

	if p, ok := s.pipes[filepath]; ok {
		return p
	}

	p := make(chan *fileOp)
	s.pipes[filepath] = p

	return p
}

func (s *Server) runPipeWorker(filepath string, pipe fileOpPipe) {
	f, err := s.fs.Open(filepath)
	defer f.Close()

	if err != nil {
		log.Panicf("cannot open %s in %v", filepath, s.fs)
	}

	doRead := func(position, maxSize uint64) ([]byte, bool, error) {
		return nil, false, nil
	}

	doWrite := func(overwrite bool, data []byte) (uint64, error) {
		return 0, nil
	}

	for {
		select {
		case op := <-pipe:
			if op.read != nil {
				r := op.read.request
				if date, end, err := doRead(r.Position, r.MaxSize); err != nil {
					op.errC <- err
				} else {
					op.read.respC <- &file.FileReadResponse{
						Filepath: filepath,
						Position: r.Position,
						Content:  date,
						EOF:      end,
					}
				}
			} else {
				r := op.write.request
				if size, err := doWrite(r.Overwrite, r.Content); err != nil {
					op.errC <- err
				} else {
					op.write.respC <- &file.FileWriteResponse{
						Filepath: filepath,
						Size:     size,
					}
				}
			}
		case <-s.stopC:
			log.Printf("stop the worker for %s", filepath)
			break
		}
	}
}
