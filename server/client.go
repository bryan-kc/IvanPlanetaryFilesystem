package server

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"io"
   "crypto/tls"
	"google.golang.org/grpc"
	"log"
	"google.golang.org/grpc/credentials"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

func (s *Server) GetRemoteFile(ctx context.Context, in *serverpb.GetRemoteFileRequest) (*serverpb.GetRemoteFileResponse, error) {
   num_hops := int32(in.NumHops)
   access_id := in.AccessId

   if num_hops > 0 {
      // for each peer, get remote file with num hops - 1
    // for _, conn := range s.mu.peerConns {
    //   // make a request and decrement num hops
    // 	fmt.Println("goign to make a request to get a doc")
    //   client := serverpb.NewClientClient(conn)
    //   args := &serverpb.GetRemoteFileRequest{
    //     AccessId: access_id,
    //     NumHops: num_hops - 1,
    //   }
    //   resp, err := client.GetRemoteFile(ctx, args)

    //  	fmt.Println("this is their response", resp)
    //   if err != nil {
    //     fmt.Println(err)
    //     return nil, err
    //   } else {
    //     if resp.GetDocument().GetContentType() == "directory" {
    //       fmt.Println("Child documents:")
    //       for k, v := range resp.GetDocument().GetChildren() {
    //         fmt.Println(k + ": " + v)
    //         return &serverpb.GetRemoteFileResponse{Document: &f}, nil
    //       }
    //     } else {
    //       fmt.Printf("%s\n", resp.Document.GetData())
    //       return &serverpb.GetRemoteFileResponse{Document: &f}, nil
    //     }
    //   }
    // }
    for _, meta := range s.mu.peerMeta {
    	creds := credentials.NewTLS(&tls.Config{
    		Rand:               rand.Reader,
    		InsecureSkipVerify: true,
    	})

    	addr := meta.Addrs[0]

    	ctx := context.TODO()
    	ctxDial, _ := context.WithTimeout(ctx, 2*time.Second)
    	conn, err := grpc.DialContext(ctxDial, addr, grpc.WithTransportCredentials(creds), grpc.WithBlock())

    	if err != nil {
    		log.Fatal(err)
    	}      
    	client := serverpb.NewClientClient(conn)
      args := &serverpb.GetRemoteFileRequest{
        AccessId: access_id,
        NumHops: num_hops - 1,
      }
      resp, err := client.GetRemoteFile(ctx, args)

      if err != nil {
        fmt.Println(err)
      } else {
        if resp.GetDocument().GetContentType() == "directory" {
          fmt.Println("Child documents:")
          for k, v := range resp.GetDocument().GetChildren() {
            fmt.Println(k + ": " + v)
            return &serverpb.GetRemoteFileResponse{Document: resp.Document}, nil
          }
        } else {
          return &serverpb.GetRemoteFileResponse{Document: resp.Document}, nil
        }
      }
    }
    fmt.Println("couldn't find the document at all.")
    return nil, errors.New("missing Document")
   } else {
      // get file and return it
    var f serverpb.Document
    documentId := strings.Split(in.GetAccessId(), ":")[0]
    parts := strings.Split(in.GetAccessId(), ":")
    if len(parts) < 2 {
      return nil, errors.Errorf("AccessId should have a :")
    }
    accessKey, err := base64.StdEncoding.DecodeString(parts[1])
    if err != nil {
      return nil, err
    }
    // Check if the file exists locally first
    if err := s.db.View(func(txn *badger.Txn) error {
      key := fmt.Sprintf("/document/%s", documentId)
      item, err := txn.Get([]byte(key))
      if err != nil {
        return err
      }
      body, err := item.Value()
      if err != nil {
        return err
      }
      if f, err = s.DecryptDocument(body, accessKey); err != nil {
        return err
      }
      return nil
    }); err != nil {
    if err != badger.ErrKeyNotFound {
      // Error wasn't an error relating to the document not being found locally. Return.
      return nil, err
    } else if err == badger.ErrKeyNotFound {
      return nil, errors.New("missing Document")
    }
  }
   resp := &serverpb.GetRemoteFileResponse{
    Document: &f,
   }
   return resp, nil
 }
}

func (s *Server) Get(ctx context.Context, in *serverpb.GetRequest) (*serverpb.GetResponse, error) {
	var f serverpb.Document
	documentId := strings.Split(in.GetAccessId(), ":")[0]
	parts := strings.Split(in.GetAccessId(), ":")
	if len(parts) < 2 {
		return nil, errors.Errorf("AccessId should have a :")
	}
	accessKey, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	// Check if the file exists locally first
	if err := s.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("/document/%s", documentId)
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		body, err := item.Value()
		if err != nil {
			return err
		}
		if f, err = s.DecryptDocument(body, accessKey); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err != badger.ErrKeyNotFound {
			// Error wasn't an error relating to the document not being found locally. Return.
			return nil, err
		} else if err == badger.ErrKeyNotFound {
			// TODO: Network check, does this document exist on the network? If so, return it (Bea)
      num_hops, _ := s.checkNumHopsToGetToFile(in.GetAccessId())

      if num_hops <= 0 {
        return nil, errors.New("missing Document")
      }
      args := &serverpb.GetRemoteFileRequest{
        AccessId: in.GetAccessId(),
        NumHops: int32(num_hops),
      }

      respRemote, err := s.GetRemoteFile(ctx,args)
      if err != nil {
        fmt.Println("Get remote file err", err)
        return nil, err
      }
      resp := &serverpb.GetResponse{
      	Document: respRemote.Document,
      }
      return resp, nil
      // make a request to get the file
			// TODO: Then, decrypt the document and return it as a Document object (Jinny or Jonathan)
		}
	}

	resp := &serverpb.GetResponse{
		Document: &f,
	}
	return resp, nil
}

func (s *Server) Add(ctx context.Context, in *serverpb.AddRequest) (*serverpb.AddResponse, error) {
	doc := in.GetDocument()
	if doc == nil {
		return nil, errors.New("missing Document")
	}
	encryptedDocument, key, err := s.EncryptDocument(*doc)
	if err != nil {
		return nil, err
	}

	data := sha1.Sum(encryptedDocument)
	hash := base64.StdEncoding.EncodeToString(data[:])

	if err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(fmt.Sprintf("/document/%s", hash)), encryptedDocument)
	}); err != nil {
		return nil, err
	}

	accessKey := base64.StdEncoding.EncodeToString(key)
	accessId := hash + ":" + accessKey
	resp := &serverpb.AddResponse{
		AccessId: accessId,
	}

  localMeta, err := s.NodeMeta()
  if err != nil {
    return nil, err
  }
  s.addToRoutingTable(localMeta.Id, accessId)

	return resp, nil
}

func (s *Server) AddDirectory(ctx context.Context, in *serverpb.AddDirectoryRequest) (*serverpb.AddDirectoryResponse, error) {
	resp := &serverpb.AddDirectoryResponse{}
	return resp, nil
}

func (s *Server) GetPeers(ctx context.Context, in *serverpb.GetPeersRequest) (*serverpb.GetPeersResponse, error) {
	var peers []*serverpb.NodeMeta
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range s.mu.peerMeta {
		peers = append(peers, &v)
	}

	resp := &serverpb.GetPeersResponse{
		Peers: peers,
	}

	return resp, nil
}

func (s *Server) AddPeer(ctx context.Context, in *serverpb.AddPeerRequest) (*serverpb.AddPeerResponse, error) {
	err := s.BootstrapAddNode(in.GetAddr())
	if err != nil {
		return nil, err
	}
	resp := &serverpb.AddPeerResponse{}
	return resp, nil
}

func (s *Server) GetReference(ctx context.Context, in *serverpb.GetReferenceRequest) (*serverpb.GetReferenceResponse, error) {
	var reference serverpb.Reference
	// Try to get reference locally first
	if err := s.db.View(func(txn *badger.Txn) error {
		key := fmt.Sprintf("/reference/%s", in.GetReferenceId())
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		body, err := item.Value()
		if err != nil {
			return err
		}
		if err = reference.Unmarshal(body); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err != badger.ErrKeyNotFound {
			// Error wasn't an error relating to the reference not being found locally. Return.
			return nil, err
		} else if err == badger.ErrKeyNotFound {
			// TODO: Do a network lookup for this reference (Bea)
		}
	}

	resp := &serverpb.GetReferenceResponse{
		Reference: &reference,
	}

	return resp, nil
}

func (s *Server) AddReference(ctx context.Context, in *serverpb.AddReferenceRequest) (*serverpb.AddReferenceResponse, error) {
	privKey, err := LoadPrivate(in.GetPrivKey())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	pubKey, err := MarshalPublic(&privKey.PublicKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// Create reference
	reference := &serverpb.Reference{
		Value:     in.GetRecord(),
		PublicKey: pubKey,
		Timestamp: time.Now().Unix(),
	}
	bytes, err := reference.Marshal()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	r, s1, err := Sign(bytes, *privKey)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	sig, err := asn1.Marshal(EcdsaSignature{R: r, S: s1})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	reference.Signature = base64.StdEncoding.EncodeToString(sig)

	// Add this reference locally
	referenceId, err := Hash(reference.PublicKey)
	if err != nil {
		return nil, err
	}
	b, err := reference.Marshal()
	if err != nil {
		return nil, err
	}

	if err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(fmt.Sprintf("/reference/%s", referenceId)), b)
	}); err != nil {
		return nil, err
	}

	// TODO: Add reference to bloom filter (Bea)

	resp := &serverpb.AddReferenceResponse{
		ReferenceId: referenceId,
	}
	return resp, nil
}

// TODO: MAKE SOME End to End behaviour testing
func (s *Server) EncryptDocument(doc serverpb.Document) (encryptedData []byte, key []byte, err error) {
	// Create a new SHA256 handler
	shaHandler := sha256.New()

	marshalledData, err := doc.Marshal()
	if err != nil {
		return nil, nil, err
	}

	// Attempt to write the data to the SHA1 handler (is this right?)
	if _, err := shaHandler.Write(marshalledData); err != nil {
		// Well, error I guess
		return nil, nil, err
	}

	// Grab the SHA1 key
	docKey := shaHandler.Sum(nil)

	// Create a new AESBlockCipher
	aesBlock, err := aes.NewCipher(docKey)
	if err != nil {
		return nil, nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(marshalledData))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	stream := cipher.NewCFBEncrypter(aesBlock, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], marshalledData)

	return ciphertext, docKey, nil
}

func (s *Server) DecryptDocument(documentData []byte, key []byte) (decryptedDocument serverpb.Document, err error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return serverpb.Document{}, err
	}

	if len(documentData) < aes.BlockSize {
		panic("ciphertext too short")
	}

	iv := documentData[:aes.BlockSize]
	documentData = documentData[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(aesBlock, iv)
	plainText := make([]byte, len(documentData))
	stream.XORKeyStream(plainText, documentData)

	err = decryptedDocument.Unmarshal(plainText)
	if err != nil {
		return serverpb.Document{}, err
	}

	return decryptedDocument, nil
}
