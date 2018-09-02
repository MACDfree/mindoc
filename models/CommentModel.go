package models

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/lifei6671/mindoc/conf"
)

//Comment struct
type Comment struct {
	CommentId int `orm:"pk;auto;unique;column(comment_id)" json:"comment_id"`
	Floor     int `orm:"column(floor);type(unsigned);default(0)" json:"floor"`
	BookId    int `orm:"column(book_id);type(int)" json:"book_id"`
	// DocumentId 评论所属的文档.
	DocumentId int `orm:"column(document_id);type(int)" json:"document_id"`
	// BlogId 评论所属的博客
	BlogId int `orm:"column(blog_id);type(int)" json:"blog_id"`
	// Author 评论作者.
	Author string `orm:"column(author);size(100)" json:"author"`
	//MemberId 评论用户ID.
	MemberId int `orm:"column(member_id);type(int)" json:"member_id"`
	// IPAddress 评论者的IP地址
	IPAddress string `orm:"column(ip_address);size(100)" json:"ip_address"`
	// 评论日期.
	CommentDate time.Time `orm:"type(datetime);column(comment_date);auto_now_add" json:"comment_date"`
	//Content 评论内容.
	Content string `orm:"column(content);size(2000)" json:"content"`
	// Approved 评论状态：0 待审核/1 已审核/2 垃圾评论/ 3 已删除
	Approved int `orm:"column(approved);type(int)" json:"approved"`
	// UserAgent 评论者浏览器内容
	UserAgent string `orm:"column(user_agent);size(500)" json:"user_agent"`
	// ParentId 评论所属父级
	ParentId     int `orm:"column(parent_id);type(int);default(0)" json:"parent_id"`
	AgreeCount   int `orm:"column(agree_count);type(int);default(0)" json:"agree_count"`
	AgainstCount int `orm:"column(against_count);type(int);default(0)" json:"against_count"`
}

// TableName 获取对应数据库表名.
func (m *Comment) TableName() string {
	return "comments"
}

// TableEngine 获取数据使用的引擎.
func (m *Comment) TableEngine() string {
	return "INNODB"
}

func (m *Comment) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewComment() *Comment {
	return &Comment{}
}
func (m *Comment) Find(id int) (*Comment, error) {
	if id <= 0 {
		return m, ErrInvalidParameter
	}
	o := orm.NewOrm()
	err := o.Read(m)

	return m, err
}

func (m *Comment) Update(cols ...string) error {
	o := orm.NewOrm()

	_, err := o.Update(m, cols...)

	return err
}

//Insert 添加一条评论.
func (m *Comment) Insert() error {
	if m.Content == "" {
		return ErrCommentContentNotEmpty
	}

	o := orm.NewOrm()

	if m.ParentId > 0 {
		comment := NewComment()
		comment.CommentId = m.ParentId
		//如果父评论不存在
		if err := o.Read(comment); err != nil {
			beego.Error(err)
			return err
		}
	}

	if m.BlogId > 0 {
		blog := NewBlog()
		if _, err := blog.Find(m.BlogId); err != nil {
			beego.Error(err)
			return err
		}
	} else {
		document := NewDocument()
		//如果评论的文档不存在
		if _, err := document.Find(m.DocumentId); err != nil {
			beego.Error(err)
			return err
		}
		book, err := NewBook().Find(document.BookId)
		//如果评论的项目不存在
		if err != nil {
			beego.Error(err)
			return err
		}
		//如果已关闭评论
		if book.CommentStatus == "closed" {
			return ErrCommentClosed
		}
		if book.CommentStatus == "registered_only" && m.MemberId <= 0 {
			return ErrPermissionDenied
		}
		//如果仅参与者评论
		if book.CommentStatus == "group_only" {
			if m.MemberId <= 0 {
				return ErrPermissionDenied
			}
			rel := NewRelationship()
			if _, err := rel.FindForRoleId(book.BookId, m.MemberId); err != nil {
				return ErrPermissionDenied
			}
		}
		m.BookId = book.BookId
	}

	if m.MemberId > 0 {
		member := NewMember()
		//如果用户不存在
		if _, err := member.Find(m.MemberId); err != nil {
			return ErrMemberNoExist
		}
		//如果用户被禁用
		if member.Status == 1 {
			return ErrMemberDisabled
		}
	} else if m.Author == "" {
		m.Author = "[匿名用户]"
	}
	_, err := o.Insert(m)
	return err
}

//分页查找标签.
func (m *Comment) FindToPager(pageIndex, pageSize int) (comments []*Comment, totalCount int, err error) {
	o := orm.NewOrm()

	count, err := o.QueryTable(m.TableNameWithPrefix()).Count()

	if err != nil {
		return
	}
	totalCount = int(count)

	offset := (pageIndex - 1) * pageSize

	_, err = o.QueryTable(m.TableNameWithPrefix()).OrderBy("comment_date").Offset(offset).Limit(pageSize).All(&comments)

	if err == orm.ErrNoRows {
		beego.Info("没有查询到标签 ->", err)
		err = nil
		return
	}
	return
}

// ListByBlogID 根据blogID获取评论记录
func (m *Comment) ListByBlogID(blogID int) (comments []*Comment, err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable(m.TableNameWithPrefix()).Filter("blog_id", blogID).OrderBy("comment_date").All(&comments)
	if err == orm.ErrNoRows {
		err = nil
		return
	}
	return
}

// ListByDocumentID 根据bookID和documentID获取评论记录
func (m *Comment) ListByDocumentID(bookID, documentID int) (comments []*Comment, err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable(m.TableNameWithPrefix()).Filter("book_id", bookID).Filter("document_id", documentID).OrderBy("-comment_date").All(&comments)
	if err == orm.ErrNoRows {
		err = nil
		return
	}
	return
}
