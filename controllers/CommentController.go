package controllers

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/lifei6671/mindoc/models"
)

// CommentController 是评论相关的后台
type CommentController struct {
	BaseController
}

// List Comments
func (c *CommentController) Lists() {
	c.Prepare()
	id, _ := c.GetInt("id")
	bookID, _ := c.GetInt("bookId")
	isBlog, _ := c.GetBool("isBlog")

	var commentDTOs []*models.CommentResult
	var comments []*models.Comment
	var err error
	// 根据blogid查找相关的评论
	if isBlog {
		comments, err = models.NewComment().ListByBlogID(id)
		beego.Info("size=", len(comments))
		if err != nil {
			beego.Error("ListComment => ", err)
			c.JsonResult(6005, "获取评论失败")
		}
	} else {
		comments, err = models.NewComment().ListByDocumentID(bookID, id)
		if err != nil {
			beego.Error("ListComment => ", err)
			c.JsonResult(6005, "获取评论失败")
		}
	}

	for _, comment := range comments {
		var parentID interface{}
		if comment.ParentId != 0 {
			parentID = strconv.Itoa(comment.ParentId)
		}
		dto := &models.CommentResult{
			ID:       strconv.Itoa(comment.CommentId),
			Author:   comment.Author,
			Comment:  comment.Content,
			ParentID: parentID,
			Date:     comment.CommentDate,
		}

		if c.Member == nil {
			dto.CanDelete = false
			dto.CanReply = true
		} else {
			dto.CanDelete = c.Member.MemberId == comment.MemberId
			dto.CanReply = c.Member.MemberId != comment.MemberId
		}

		commentDTOs = append(commentDTOs, dto)
	}

	returnJSON, err := json.Marshal(commentDTOs)
	if err != nil {
		beego.Error(err)
	}

	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache, no-store")
	io.WriteString(c.Ctx.ResponseWriter, string(returnJSON))

	c.StopRun()
}

func (c *CommentController) Index() {
	c.Prepare()
	c.TplName = "comment/index.tpl"
}

func (c *CommentController) Create() {
	c.Prepare()
	id, _ := c.GetInt("id")
	beego.Info("id=", id)
	bookID, _ := c.GetInt("bookId")
	isBlog, _ := c.GetBool("isBlog")
	parentID, _ := c.GetInt("parentId")
	content := c.GetString("comment")

	comment := models.NewComment()
	comment.Content = content
	comment.ParentId = parentID
	if c.Member != nil {
		comment.Author = c.Member.Account
		comment.MemberId = c.Member.MemberId
	}
	comment.Approved = 1 // 默认已审核
	if isBlog {
		comment.BlogId = id
	} else {
		comment.BookId = bookID
		comment.DocumentId = id
	}
	err := comment.Insert()
	if err != nil {
		beego.Error(err)
	}

	var parent interface{}
	if comment.ParentId != 0 {
		parent = strconv.Itoa(comment.ParentId)
	}
	dto := &models.CommentResult{
		ID:       strconv.Itoa(comment.CommentId),
		Author:   comment.Author,
		Comment:  comment.Content,
		ParentID: parent,
		Date:     comment.CommentDate,
	}

	if c.Member == nil {
		dto.CanDelete = false
		dto.CanReply = true
	} else {
		dto.CanDelete = c.Member.MemberId == comment.MemberId
		dto.CanReply = c.Member.MemberId != comment.MemberId
	}

	returnJSON, err := json.Marshal(dto)
	if err != nil {
		beego.Error(err)
	}

	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache, no-store")
	io.WriteString(c.Ctx.ResponseWriter, string(returnJSON))

	c.StopRun()
}
